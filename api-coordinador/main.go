package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"api-coordinador/internal/analisis"
	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
	"api-coordinador/internal/models"

	"context"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/golang-jwt/jwt/v5"
)

// Estado Global Protegido
var (
	bosqueGlobal []*models.TreeNode
	matrizGlobal [3][3]int
	rwMutex      sync.RWMutex
	jwtSecret    = []byte("secreto-super-seguro-pc4")
	mongoClient  *mongo.Client
	historialCol *mongo.Collection
)

type Credenciales struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
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

	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/train", JWTMiddleware(handleTrain))
	http.HandleFunc("/api/predict", handlePredict)
	http.HandleFunc("/api/metrics", JWTMiddleware(handleMetrics))

	fmt.Println("[API-REST] Servidor HTTP de escucha perpetua iniciado en :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("[CRÍTICO] Fallo en el servidor HTTP: %v\n", err)
	}
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

	if creds.Username != "admin" || creds.Password != "admin123" {
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

	// Evaluación centralizada Map-Reduce
	nuevaMatriz := analisis.EvaluarBosqueDistribuido(testDataRaw, nuevoBosque, numWorkers)
	rwMutex.Lock()
	bosqueGlobal = nuevoBosque
	matrizGlobal = nuevaMatriz
	rwMutex.Unlock()

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

	var p models.PerfilPaciente
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}

	// Inferencia con Bloqueo Compartido
	rwMutex.RLock()
	bosqueLocal := bosqueGlobal
	rwMutex.RUnlock()

	if len(bosqueLocal) == 0 {
		http.Error(w, "El modelo aún no ha sido entrenado", http.StatusServiceUnavailable)
		return
	}

	clase := analisis.PredecirRandomForest(p, bosqueLocal)

	// Persistencia Asíncrona (Fire & Forget)
	go guardarHistorialEnMongo(p, clase)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint8{"prediction": clase})
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
