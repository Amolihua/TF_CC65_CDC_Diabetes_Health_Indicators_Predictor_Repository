package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"api-coordinador/internal/analisis"
	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
	"api-coordinador/internal/models"

	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Estructuras Singleflight
type sfCall struct {
	wg  sync.WaitGroup
	val uint8
}

// Estado Global Protegido
var (
	bosqueGlobal []*models.TreeNode
	matrizGlobal [3][3]int
	rwMutex      sync.RWMutex
	jwtSecret    []byte
	mongoClient  *mongo.Client
	historialCol *mongo.Collection
	rdb          *redis.Client

	// Observabilidad y Caché
	modeloVersion uint64 = 1
	cacheHits     uint64
	cacheMisses   uint64
	cacheErrors   uint64
	activeSockets uint64

	// Estado Singleflight
	sfGroup = make(map[string]*sfCall)
	sfMutex sync.Mutex
)

type Credenciales struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secreto-super-seguro-pc4" // Fallback
	}
	jwtSecret = []byte(secret)

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("[CRÍTICO] Fallo al conectar con MongoDB: %v\n", err)
	} else {
		mongoClient = client
		historialCol = client.Database("cdc_diabetes").Collection("predicciones")
		fmt.Println("[API-REST] Conexión establecida con MongoDB en", mongoURI)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("[API-REST] ADVERTENCIA: Redis inaccesible en %s: %v\n", redisAddr, err)
	} else {
		fmt.Printf("[API-REST] Ping exitoso a Redis en %s\n", redisAddr)
	}

	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/train", JWTMiddleware(handleTrain))
	http.HandleFunc("/api/predict", handlePredict)
	http.HandleFunc("/api/metrics", JWTMiddleware(handleMetrics))
	http.HandleFunc("/api/ws/metrics", handleWSMetrics)

	fmt.Println("[API-REST] Servidor HTTP de escucha perpetua iniciado en :8080")
	if err := http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux)); err != nil {
		fmt.Printf("[CRÍTICO] Fallo en el servidor HTTP: %v\n", err)
	}
}

// Inyectar CORS a las respuestas
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Intercepta peticiones, extrae Bearer Token y verifica la expiración
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"No autorizado"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Firma inesperada")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Token inválido o expirado"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Endpoint público para expedir token con 24h de expiración
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var creds Credenciales
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Petición inválida", http.StatusBadRequest)
		return
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123"
	}

	if creds.Username != adminUser || creds.Password != adminPass {
		http.Error(w, `{"error":"Credenciales incorrectas"}`, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error generando token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func handleTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	inicio := time.Now()
	numWorkers := leerEnteroEnv("NUM_WORKERS", 12)

	r.Body = http.MaxBytesReader(w, r.Body, 400<<20) // 400 MB Límite
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Error al procesar multipart", http.StatusBadRequest)
		return
	}

	var filePart io.Reader
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error leyendo partes", http.StatusInternalServerError)
			return
		}
		if part.FormName() == "dataset" {
			filePart = part
			break
		}
	}

	if filePart == nil {
		http.Error(w, "Archivo dataset no encontrado", http.StatusBadRequest)
		return
	}

	nodosStr := os.Getenv("NODOS_ML_ADDRS")
	if nodosStr == "" {
		nodosStr = "localhost:9000"
	}
	nodosAddrs := strings.Split(nodosStr, ",")
	numNodos := len(nodosAddrs)

	// Pipeline de ingesta y partición
	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(filePart, jobs)
	canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

	var writers []*bufio.Writer
	var conns []net.Conn
	baseTrees := 50 / numNodos
	remainder := 50 % numNodos

	// Conexión a nodos esclavos TCP
	for i, addr := range nodosAddrs {
		conn, err := net.Dial("tcp", strings.TrimSpace(addr))
		if err != nil {
			continue
		}
		conns = append(conns, conn)
		writer := bufio.NewWriterSize(conn, 256*1024)

		treesPerNode := baseTrees
		if i < remainder {
			treesPerNode++
		}
		fmt.Fprintf(writer, `{"algoritmo":"random_forest","num_workers":%d,"num_trees":%d}`+"\n", numWorkers, treesPerNode)
		writers = append(writers, writer)
	}

	if len(writers) == 0 {
		http.Error(w, "No hay nodos ML disponibles", http.StatusInternalServerError)
		return
	}

	var testDataRaw [][]byte
	count, nodeIndex := 0, 0

	// Sharding y 20% retención local
	for jsonBytes := range canalLimpio {
		if count%10 < 8 {
			writer := writers[nodeIndex]
			_, _ = writer.Write(jsonBytes)
			_ = writer.WriteByte('\n')
			nodeIndex = (nodeIndex + 1) % len(writers)
		} else {
			clone := make([]byte, len(jsonBytes))
			copy(clone, jsonBytes)
			testDataRaw = append(testDataRaw, clone)
		}
		count++
	}

	// Cerrar flujos TCP
	for i, writer := range writers {
		_ = writer.Flush()
		if tcpConn, ok := conns[i].(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}

	var wg sync.WaitGroup
	var nuevoBosque []*models.TreeNode
	var mu sync.Mutex

	// Recepción binaria y ensamblaje concurrente
	for _, conn := range conns {
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer c.Close()
			data, err := io.ReadAll(c)
			if err != nil && err != io.EOF {
				return
			}
			subBosque := models.DeserializeForest(data)
			mu.Lock()
			nuevoBosque = append(nuevoBosque, subBosque...)
			mu.Unlock()
		}(conn)
	}
	wg.Wait()

	// Validación integridad distribuida
	if len(nuevoBosque) != 50 {
		errMsg := fmt.Sprintf("Error de integridad en el clúster: Se esperaban 50 árboles, pero los nodos devolvieron %d. Entrenamiento abortado.", len(nuevoBosque))
		fmt.Printf("[CRÍTICO] %s\n", errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	fmt.Printf("[API-REST] Integridad validada: Se recibieron exactamente %d árboles de los nodos esclavos.\n", len(nuevoBosque))

	// Evaluación centralizada Map-Reduce
	nuevaMatriz := analisis.EvaluarBosqueDistribuido(testDataRaw, nuevoBosque, numWorkers)
	rwMutex.Lock()
	bosqueGlobal = nuevoBosque
	matrizGlobal = nuevaMatriz
	rwMutex.Unlock()

	// Incremento atómico de versión
	atomic.AddUint64(&modeloVersion, 1)

	tiempoTotal := time.Since(inicio).String()
	fmt.Printf("[API-REST] Entrenamiento finalizado. Tiempo: %s\n", tiempoTotal)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "Entrenamiento finalizado exitosamente",
		"time_elapsed": tiempoTotal,
	})
}

func handlePredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	inicio := time.Now()

	var p models.PerfilPaciente
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("pred:v%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%f:%f:%d:%d:%d:%d:%d",
		atomic.LoadUint64(&modeloVersion),
		p.HighBP, p.HighChol, p.CholCheck, p.BMI, p.Smoker, p.Stroke, p.HeartDiseaseorAttack,
		p.PhysActivity, p.Fruits, p.Veggies, p.HvyAlcoholConsump, p.AnyHealthcare, p.NoDocbcCost,
		p.GenHlth, p.MentHlth, p.PhysHlth, p.DiffWalk, p.Sex, p.Age, p.Education, p.Income)

	ctxRedisGet, cancelGet := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelGet()

	val, err := rdb.Get(ctxRedisGet, key).Result()
	if err == nil {
		atomic.AddUint64(&cacheHits, 1)
		clase, _ := strconv.Atoi(val)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"prediction":   uint8(clase),
			"from_cache":   true,
			"time_elapsed": time.Since(inicio).String(),
		})
		return
	} else if err == redis.Nil {
		atomic.AddUint64(&cacheMisses, 1)
	} else {
		atomic.AddUint64(&cacheErrors, 1)
		fmt.Printf("[API-REST] ADVERTENCIA: Error en caché obteniendo clave %s: %v\n", key, err)
	}

	// Sincronización Singleflight artesanal
	sfMutex.Lock()
	if c, ok := sfGroup[key]; ok {
		sfMutex.Unlock()
		c.wg.Wait() // Esperar a la petición líder
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"prediction":   c.val,
			"from_cache":   false,
			"time_elapsed": time.Since(inicio).String(),
		})
		return
	}
	c := new(sfCall)
	c.wg.Add(1)
	sfGroup[key] = c
	sfMutex.Unlock()

	// Inferencia con Bloqueo Compartido
	rwMutex.RLock()
	bosqueLocal := bosqueGlobal
	rwMutex.RUnlock()

	var clase uint8
	if len(bosqueLocal) == 0 {
		http.Error(w, "El modelo aún no ha sido entrenado", http.StatusServiceUnavailable)
		// Liberar Singleflight en error
		sfMutex.Lock()
		delete(sfGroup, key)
		sfMutex.Unlock()
		c.wg.Done()
		return
	}

	clase = analisis.PredecirRandomForest(p, bosqueLocal)

	// Compartir resultado Singleflight y liberar
	c.val = clase
	sfMutex.Lock()
	delete(sfGroup, key)
	sfMutex.Unlock()
	c.wg.Done()

	// Persistencia Asíncrona Combinada (Caché + MongoDB)
	go func(llave string, valor uint8, perfil models.PerfilPaciente) {
		ctxRedisSet, cancelSet := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancelSet()
		rdb.Set(ctxRedisSet, llave, valor, 12*time.Hour)
		guardarHistorialEnMongo(perfil, valor)
	}(key, clase, p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prediction":   clase,
		"from_cache":   false,
		"time_elapsed": time.Since(inicio).String(),
	})
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Lectura de métricas con Bloqueo Compartido
	rwMutex.RLock()
	matriz := matrizGlobal
	rwMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"confusion_matrix": matriz,
		"cache_hits":       atomic.LoadUint64(&cacheHits),
		"cache_misses":     atomic.LoadUint64(&cacheMisses),
		"cache_errors":     atomic.LoadUint64(&cacheErrors),
		"modelo_version":   atomic.LoadUint64(&modeloVersion),
	})
}

func leerEnteroEnv(nombre string, valorDefault int) int {
	valor, err := strconv.Atoi(os.Getenv(nombre))
	if err != nil || valor <= 0 {
		return valorDefault
	}
	return valor
}

func guardarHistorialEnMongo(p models.PerfilPaciente, diagnosis uint8) {
	if historialCol == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := models.HistorialPredictivo{
		Perfil:    p,
		Diagnosis: diagnosis,
		CreatedAt: time.Now(),
	}
	_, _ = historialCol.InsertOne(ctx, doc)
}

func handleWSMetrics(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	conn, bufrw, _ := w.(http.Hijacker).Hijack()
	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n")
	bufrw.Flush()

	go func() {
		atomic.AddUint64(&activeSockets, 1)
		defer atomic.AddUint64(&activeSockets, ^uint64(0))
		defer conn.Close()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			rwMutex.RLock()
			matriz := matrizGlobal
			rwMutex.RUnlock()

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			payload, _ := json.Marshal(map[string]interface{}{
				"cache_hits":     atomic.LoadUint64(&cacheHits),
				"cache_misses":   atomic.LoadUint64(&cacheMisses),
				"cache_errors":   atomic.LoadUint64(&cacheErrors),
				"modelo_version": atomic.LoadUint64(&modeloVersion),
				"nodos_activos":  atomic.LoadUint64(&activeSockets),
				"cpu_goroutines": runtime.NumGoroutine(),
				"ram_sys_mb":     m.Sys / 1024 / 1024,
				"matriz":         matriz,
			})

			size := len(payload)
			var header []byte
			if size <= 125 {
				header = []byte{0x81, byte(size)}
			} else {
				header = []byte{0x81, 126, byte(size >> 8), byte(size & 255)}
			}

			// Intercepción de error --> Desconexión limpia del cliente
			if _, err := conn.Write(append(header, payload...)); err != nil {
				return
			}
		}
	}()
}
