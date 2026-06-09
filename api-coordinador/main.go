package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
)

func main() {
	inicio := time.Now()

	numWorkers := leerEnteroEnv("NUM_WORKERS", 12)
	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	nodoAddr := os.Getenv("NODO_ML_ADDR")
	if nodoAddr == "" {
		nodoAddr = "localhost:9000"
	}

	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(datasetPath, jobs)
	canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

	conn, _ := net.Dial("tcp", nodoAddr)
	defer conn.Close()

	algoritmoAEntrenar := "random_forest"
	bufWriter := bufio.NewWriterSize(conn, 256*1024)
	fmt.Fprintf(bufWriter, `{"algoritmo":%q,"num_workers":%d}`+"\n", algoritmoAEntrenar, numWorkers)

	count := 0
	for jsonBytes := range canalLimpio {
		_, _ = bufWriter.Write(jsonBytes)
		_ = bufWriter.WriteByte('\n')
		count++
	}

	_ = bufWriter.Flush()
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.CloseWrite()
	}

	_, _ = io.ReadAll(conn)
	fmt.Printf("[API-COORDINADOR] Dataset de %d registros enviado a %s en: %s\n", count, nodoAddr, time.Since(inicio))
}

func leerEnteroEnv(nombre string, valorDefault int) int {
	valor, err := strconv.Atoi(os.Getenv(nombre))
	if err != nil || valor <= 0 {
		return valorDefault
	}
	return valor
}
