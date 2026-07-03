package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"api-coordinador/internal/analisis"
	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
	"api-coordinador/internal/models"

	"github.com/redis/go-redis/v9"
)

func (s *Server) HandleTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	inicio := time.Now()
	numWorkers := leerEnteroEnv("NUM_WORKERS", 12)

	r.Body = http.MaxBytesReader(w, r.Body, 400<<20)
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

	// Pipeline
	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(filePart, jobs)
	canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

	var writers []*bufio.Writer
	var conns []net.Conn
	baseTrees := 50 / numNodos
	remainder := 50 % numNodos

	// Conexión TCP
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

	// Sharding
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

	// Recepción binaria y ensamblaje
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

	// Validación
	if len(nuevoBosque) != 50 {
		errMsg := fmt.Sprintf("Error de integridad en el clúster: Se esperaban 50 árboles, pero los nodos devolvieron %d. Entrenamiento abortado.", len(nuevoBosque))
		fmt.Printf("[CRÍTICO] %s\n", errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	fmt.Printf("[API-REST] Integridad validada: Se recibieron exactamente %d árboles de los nodos esclavos.\n", len(nuevoBosque))

	// Evaluación centralizada Map-Reduce
	nuevaMatriz := analisis.EvaluarBosqueDistribuido(testDataRaw, nuevoBosque, numWorkers)
	s.RWMutex.Lock()
	s.BosqueGlobal = nuevoBosque
	s.MatrizGlobal = nuevaMatriz
	s.RWMutex.Unlock()

	// Incremento atómico de versión
	atomic.AddUint64(&s.ModeloVersion, 1)

	tiempoTotal := time.Since(inicio).String()
	fmt.Printf("[API-REST] Entrenamiento finalizado. Tiempo: %s\n", tiempoTotal)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "Entrenamiento finalizado exitosamente",
		"time_elapsed": tiempoTotal,
	})
}

func (s *Server) HandlePredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	inicio := time.Now()

	var req struct {
		models.PerfilPaciente
		Email string `json:"email,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}
	p := req.PerfilPaciente

	key := fmt.Sprintf("pred:v%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%f:%f:%d:%d:%d:%d:%d",
		atomic.LoadUint64(&s.ModeloVersion),
		p.HighBP, p.HighChol, p.CholCheck, p.BMI, p.Smoker, p.Stroke, p.HeartDiseaseorAttack,
		p.PhysActivity, p.Fruits, p.Veggies, p.HvyAlcoholConsump, p.AnyHealthcare, p.NoDocbcCost,
		p.GenHlth, p.MentHlth, p.PhysHlth, p.DiffWalk, p.Sex, p.Age, p.Education, p.Income)

	ctxRedisGet, cancelGet := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelGet()

	val, err := s.Rdb.Get(ctxRedisGet, key).Result()
	if err == nil {
		atomic.AddUint64(&s.CacheHits, 1)
		clase, _ := strconv.Atoi(val)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"prediction":   uint8(clase),
			"from_cache":   true,
			"time_elapsed": time.Since(inicio).String(),
		})
		return
	} else if err == redis.Nil {
		atomic.AddUint64(&s.CacheMisses, 1)
	} else {
		atomic.AddUint64(&s.CacheErrors, 1)
		fmt.Printf("[API-REST] ADVERTENCIA: Error en caché obteniendo clave %s: %v\n", key, err)
	}

	// Sincronización Singleflight
	s.SfMutex.Lock()
	if c, ok := s.SfGroup[key]; ok {
		s.SfMutex.Unlock()
		c.wg.Wait()
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
	s.SfGroup[key] = c
	s.SfMutex.Unlock()

	// Inferencia con Bloqueo Compartido
	s.RWMutex.RLock()
	bosqueLocal := s.BosqueGlobal
	s.RWMutex.RUnlock()

	var clase uint8
	if len(bosqueLocal) == 0 {
		http.Error(w, "El modelo aún no ha sido entrenado", http.StatusServiceUnavailable)

		s.SfMutex.Lock()
		delete(s.SfGroup, key)
		s.SfMutex.Unlock()
		c.wg.Done()
		return
	}

	clase = analisis.PredecirRandomForest(p, bosqueLocal)

	c.val = clase
	s.SfMutex.Lock()
	delete(s.SfGroup, key)
	s.SfMutex.Unlock()
	c.wg.Done()

	// Persistencia Asíncrona Combinada
	go func(llave string, valor uint8, perfil models.PerfilPaciente, email string) {
		ctxRedisSet, cancelSet := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancelSet()
		s.Rdb.Set(ctxRedisSet, llave, valor, 12*time.Hour)
		s.GuardarHistorialEnMongo(perfil, valor, email)
	}(key, clase, p, req.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prediction":   clase,
		"from_cache":   false,
		"time_elapsed": time.Since(inicio).String(),
	})
}

func leerEnteroEnv(nombre string, valorDefault int) int {
	valor, err := strconv.Atoi(os.Getenv(nombre))
	if err != nil || valor <= 0 {
		return valorDefault
	}
	return valor
}
