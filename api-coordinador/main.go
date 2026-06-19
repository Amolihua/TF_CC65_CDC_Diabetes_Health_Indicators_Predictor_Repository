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
)

// Estado Global Protegido
var (
	bosqueGlobal []*models.TreeNode
	matrizGlobal [3][3]int
	rwMutex      sync.RWMutex
)

func main() {
	http.HandleFunc("/api/train", handleTrain)
	http.HandleFunc("/api/predict", handlePredict)
	http.HandleFunc("/api/metrics", handleMetrics)

	fmt.Println("[API-REST] Servidor HTTP de escucha perpetua iniciado en :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("[CRÍTICO] Fallo en el servidor HTTP: %v\n", err)
	}
}

func handleTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	inicio := time.Now()
	numWorkers := leerEnteroEnv("NUM_WORKERS", 12)
	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	nodosStr := os.Getenv("NODOS_ML_ADDRS")
	if nodosStr == "" {
		nodosStr = "localhost:9000"
	}
	nodosAddrs := strings.Split(nodosStr, ",")
	numNodos := len(nodosAddrs)

	// Pipeline de ingesta y partición
	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(datasetPath, jobs)
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
