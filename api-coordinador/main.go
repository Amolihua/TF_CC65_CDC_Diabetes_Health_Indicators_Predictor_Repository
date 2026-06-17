package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
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

func main() {
	inicio := time.Now()

	numWorkers := leerEnteroEnv("NUM_WORKERS", 12)
	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	// 1. Decoupling de Red
	nodosStr := os.Getenv("NODOS_ML_ADDRS")
	if nodosStr == "" {
		nodosStr = "localhost:9000"
	}
	nodosAddrs := strings.Split(nodosStr, ",")
	numNodos := len(nodosAddrs)
	fmt.Printf("[API-COORDINADOR] Iniciando clúster con %d nodos.\n", numNodos)

	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(datasetPath, jobs)
	canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

	// 2. Conectar a los Nodos
	var writers []*bufio.Writer
	var conns []net.Conn

	// Calcular árboles por nodo con distribución exacta
	baseTrees := 50 / numNodos
	remainder := 50 % numNodos

	for i, addr := range nodosAddrs {
		conn, err := net.Dial("tcp", strings.TrimSpace(addr))
		if err != nil {
			fmt.Printf("[ERROR] No se pudo conectar a %s: %v\n", addr, err)
			continue
		}
		conns = append(conns, conn)
		w := bufio.NewWriterSize(conn, 256*1024)

		treesPerNode := baseTrees
		if i < remainder {
			treesPerNode++
		}

		// Metadatos iniciales
		fmt.Fprintf(w, `{"algoritmo":"random_forest","num_workers":%d,"num_trees":%d}`+"\n", numWorkers, treesPerNode)
		writers = append(writers, w)
	}

	if len(writers) == 0 {
		panic("[CRÍTICO] No hay nodos disponibles. Abortando.")
	}

	var testDataRaw [][]byte
	count := 0
	nodeIndex := 0

	// 3. Sharding Dinámico y Split 80/20 Al Vuelo
	for jsonBytes := range canalLimpio {
		if count%10 < 8 {
			w := writers[nodeIndex]
			_, _ = w.Write(jsonBytes)
			_ = w.WriteByte('\n')
			nodeIndex = (nodeIndex + 1) % len(writers)
		} else {
			clone := make([]byte, len(jsonBytes))
			copy(clone, jsonBytes)
			testDataRaw = append(testDataRaw, clone)
		}
		count++
	}

	// Cerrar flujos de escritura
	for i, w := range writers {
		_ = w.Flush()
		if tcpConn, ok := conns[i].(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}

	fmt.Printf("[API-COORDINADOR] Datos distribuidos (Total: %d, Test: %d). Esperando modelos binarios...\n", count, len(testDataRaw))

	// 4. Ensamblaje Asíncrono
	var wg sync.WaitGroup
	var bosqueGlobal []*models.TreeNode
	var mu sync.Mutex

	for i, conn := range conns {
		wg.Add(1)
		go func(c net.Conn, nodeID int) {
			defer wg.Done()
			defer c.Close()

			// Extraer toda la carga binaria devuelta por el nodo
			data, err := io.ReadAll(c)
			if err != nil && err != io.EOF {
				fmt.Printf("[ERROR] Leyendo del nodo %d: %v\n", nodeID, err)
				return
			}

			// Deserializar el bosque de este esclavo
			subBosque := models.DeserializeForest(data)

			// Exclusión Mutua
			mu.Lock()
			bosqueGlobal = append(bosqueGlobal, subBosque...)
			mu.Unlock()

			fmt.Printf("[API-COORDINADOR] Recibidos %d árboles del Nodo %s\n", len(subBosque), nodosAddrs[nodeID])
		}(conn, i)
	}

	wg.Wait()
	fmt.Printf("[API-COORDINADOR] Bosque global ensamblado exitosamente con %d árboles.\n", len(bosqueGlobal))

	// 5. Evaluación Centralizada
	analisis.EvaluarBosqueDistribuido(testDataRaw, bosqueGlobal, numWorkers)

	fmt.Printf("\n[API-COORDINADOR] Pipeline PC4 Finalizado. Tiempo Total: %s\n", time.Since(inicio))
}

func leerEnteroEnv(nombre string, valorDefault int) int {
	valor, err := strconv.Atoi(os.Getenv(nombre))
	if err != nil || valor <= 0 {
		return valorDefault
	}
	return valor
}
