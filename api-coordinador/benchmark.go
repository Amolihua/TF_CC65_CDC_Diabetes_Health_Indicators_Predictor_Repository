//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"sort"
	"time"

	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
)

func main() {
	benchmarkInicio := time.Now()

	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	nodoAddr := os.Getenv("NODO_ML_ADDR")
	if nodoAddr == "" {
		nodoAddr = "localhost:9000"
	}

	algoritmos := []string{"softmax", "random_forest", "naive_bayes"}
	workerCounts := []int{2, 4, 8, 12, 16}
	iteraciones := 10

	fmt.Println(">> Iniciando Benchmark")
	fmt.Printf(">> Dataset: %s\n", datasetPath)
	fmt.Printf(">> Nodo ML: %s\n", nodoAddr)

	for _, algoritmo := range algoritmos {
		fmt.Printf("\n======================================================\n")
		fmt.Printf("EVALUANDO ALGORITMO: %s\n", algoritmo)
		fmt.Printf("======================================================\n")

		for _, numWorkers := range workerCounts {
			var tiempos []time.Duration
			var tiemposEnvio []time.Duration
			var tiemposNodo []time.Duration
			var memHeapAlloc uint64
			var memTotalAlloc uint64
			var gcTotal uint32
			var registrosProcesados int
			var maxGoroutines int

			fmt.Printf("\n  -> Prueba con %d Goroutines (10 iteraciones)...\n", numWorkers)

			for i := 0; i < iteraciones; i++ {
				var memInicio runtime.MemStats
				runtime.ReadMemStats(&memInicio)

				inicio := time.Now()

				jobs := make(chan []string, 10000)
				go loader.LeerCSVMasivo(datasetPath, jobs)
				canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

				conn, err := net.Dial("tcp", nodoAddr)
				if err != nil {
					fmt.Printf("     [!] Error de red: %v\n", err)
					continue
				}

				bufWriter := bufio.NewWriterSize(conn, 256*1024)
				fmt.Fprintf(bufWriter, `{"algoritmo":%q,"num_workers":%d}`+"\n", algoritmo, numWorkers)

				count := 0
				for jsonBytes := range canalLimpio {
					_, _ = bufWriter.Write(jsonBytes)
					_ = bufWriter.WriteByte('\n')
					count++

					if i == 0 && runtime.NumGoroutine() > maxGoroutines {
						maxGoroutines = runtime.NumGoroutine()
					}
				}

				_ = bufWriter.Flush()
				if tcpConn, ok := conn.(*net.TCPConn); ok {
					_ = tcpConn.CloseWrite()
				}

				tiempoEnvio := time.Since(inicio)
				inicioEsperaNodo := time.Now()
				_, _ = io.ReadAll(conn)
				tiempoNodo := time.Since(inicioEsperaNodo)

				duracion := time.Since(inicio)
				_ = conn.Close()

				tiempos = append(tiempos, duracion)
				tiemposEnvio = append(tiemposEnvio, tiempoEnvio)
				tiemposNodo = append(tiemposNodo, tiempoNodo)
				registrosProcesados = count

				var memFin runtime.MemStats
				runtime.ReadMemStats(&memFin)
				memHeapAlloc += memFin.Alloc
				memTotalAlloc += memFin.TotalAlloc - memInicio.TotalAlloc
				gcTotal += memFin.NumGC - memInicio.NumGC

				fmt.Printf("     - Iter %d: Total=%s | Envio=%s | Nodo=%s\n", i+1, duracion, tiempoEnvio, tiempoNodo)
			}

			if len(tiempos) < 3 {
				fmt.Println("     [!] No hubo suficientes iteraciones exitosas.")
				continue
			}

			sort.Slice(tiempos, func(i, j int) bool { return tiempos[i] < tiempos[j] })

			var suma time.Duration
			for i := 1; i < len(tiempos)-1; i++ {
				suma += tiempos[i]
			}

			mediaRecortada := suma / time.Duration(len(tiempos)-2)
			mediaEnvio := promedio(tiemposEnvio)
			mediaNodo := promedio(tiemposNodo)
			mediaHeapMB := float64(memHeapAlloc/uint64(len(tiempos))) / (1024 * 1024)
			mediaTotalAllocMB := float64(memTotalAlloc/uint64(len(tiempos))) / (1024 * 1024)
			mediaGC := gcTotal / uint32(len(tiempos))

			fmt.Printf("METRICA -> Algoritmo: %s | Workers_Coord: %d | Workers_Nodo: %d | Tiempo: %s | Envio_Prom: %s | Nodo_Prom: %s | RAM_Heap: %.2f MB | RAM_Alloc_Delta: %.2f MB | Pico_Goroutines: %d | Ciclos_GC_Delta: %d | CPUs: %d | Total_Registros: %d\n",
				algoritmo, numWorkers, numWorkers, mediaRecortada, mediaEnvio, mediaNodo, mediaHeapMB, mediaTotalAllocMB, maxGoroutines, mediaGC, runtime.NumCPU(), registrosProcesados)
		}
	}

	fmt.Printf("\n>> Benchmark finalizado en %s.\n", time.Since(benchmarkInicio))
}

func promedio(valores []time.Duration) time.Duration {
	if len(valores) == 0 {
		return 0
	}
	var suma time.Duration
	for _, valor := range valores {
		suma += valor
	}
	return suma / time.Duration(len(valores))
}
