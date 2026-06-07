//go:build ignore

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"nodo-ml/internal/engine"
	"nodo-ml/internal/models"
)

type MetadataEntrenamiento struct {
	Algoritmo      string `json:"algoritmo"`
	NumWorkers     int    `json:"num_workers"`
	TotalRegistros int    `json:"total_registros,omitempty"`
}

type ResultadoCarga struct {
	Meta                 MetadataEntrenamiento
	Dataset              []models.PerfilPaciente
	TiempoRecepcionParse time.Duration
}

func main() {
	puerto := os.Getenv("TCP_PORT")
	if puerto == "" {
		puerto = "9000"
	}

	listener, _ := net.Listen("tcp", ":"+puerto)
	fmt.Printf("Nodo ML Benchmark escuchando en puerto: |-| %s |-| \n", puerto)

	for {
		conn, _ := listener.Accept()
		go manejarConexionBenchmark(conn)
	}
}

func manejarConexionBenchmark(conn net.Conn) {
	defer conn.Close()

	inicioTotal := time.Now()
	resultadoCarga := recibirDatasetBenchmark(conn)
	numWorkers := resolverWorkers(resultadoCarga.Meta.NumWorkers)

	inicioEntrenamiento := time.Now()
	entrenar(resultadoCarga.Meta.Algoritmo, resultadoCarga.Dataset, numWorkers)
	tiempoEntrenamiento := time.Since(inicioEntrenamiento)
	tiempoTotal := time.Since(inicioTotal)

	fmt.Printf("NODO_METRICA -> Algoritmo: %s | Workers_Nodo: %d | Registros: %d | Recepcion_Parse: %s | Entrenamiento: %s | Total_Nodo: %s | Goroutines: %d | CPUs: %d\n",
		resultadoCarga.Meta.Algoritmo, numWorkers, len(resultadoCarga.Dataset), resultadoCarga.TiempoRecepcionParse, tiempoEntrenamiento, tiempoTotal, runtime.NumGoroutine(), runtime.NumCPU())

	fmt.Fprintf(conn, "OK registros=%d workers=%d recepcion_parse=%s entrenamiento=%s total_nodo=%s\n",
		len(resultadoCarga.Dataset), numWorkers, resultadoCarga.TiempoRecepcionParse, tiempoEntrenamiento, tiempoTotal)
}

func recibirDatasetBenchmark(conn net.Conn) ResultadoCarga {
	inicio := time.Now()
	var meta MetadataEntrenamiento
	var dataset []models.PerfilPaciente

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, 64*1024*1024)

	if scanner.Scan() {
		_ = json.Unmarshal(scanner.Bytes(), &meta)
	}

	if meta.TotalRegistros > 0 {
		dataset = make([]models.PerfilPaciente, 0, meta.TotalRegistros)
	}

	numWorkers := resolverWorkers(meta.NumWorkers)
	rawJobs := make(chan []byte, 10000)
	parsedResults := make(chan models.PerfilPaciente, 10000)
	var wg sync.WaitGroup
	var bytePool sync.Pool

	bytePool.New = func() any {
		buf := make([]byte, 0, 1024)
		return &buf
	}

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for raw := range rawJobs {
				parsedResults <- parsearPerfilCompacto(raw)
				if cap(raw) <= 64*1024 {
					raw = raw[:0]
					bytePool.Put(&raw)
				}
			}
		}()
	}

	done := make(chan bool)
	go func() {
		for p := range parsedResults {
			dataset = append(dataset, p)
		}
		done <- true
	}()

	for scanner.Scan() {
		linea := scanner.Bytes()
		if len(linea) < 2 {
			continue
		}

		bufPtr := bytePool.Get().(*[]byte)
		clon := append((*bufPtr)[:0], linea...)
		rawJobs <- clon
	}

	close(rawJobs)
	wg.Wait()
	close(parsedResults)
	<-done

	return ResultadoCarga{
		Meta:                 meta,
		Dataset:              dataset,
		TiempoRecepcionParse: time.Since(inicio),
	}
}

func entrenar(algoritmo string, dataset []models.PerfilPaciente, numWorkers int) {
	chunkSize := len(dataset) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	switch algoritmo {
	case "softmax":
		chunks := particionarDatos(dataset, chunkSize)
		engine.EntrenarSoftmax(chunks, 0.005, 150)
	case "random_forest":
		engine.EntrenarRandomForest(dataset, 50, numWorkers)
	case "naive_bayes":
		chunks := particionarDatos(dataset, chunkSize)
		engine.EntrenarNaiveBayes(chunks)
	default:
		fmt.Printf("[WARN] Algoritmo no soportado: %s\n", algoritmo)
	}
}

func resolverWorkers(valor int) int {
	if valor > 0 {
		return valor
	}
	if envWorkers, err := strconv.Atoi(os.Getenv("NUM_WORKERS")); err == nil && envWorkers > 0 {
		return envWorkers
	}
	return runtime.NumCPU()
}

func parsearPerfilCompacto(linea []byte) models.PerfilPaciente {
	campos := bytes.Split(linea, []byte{','})
	return models.PerfilPaciente{
		Diabetes012:          parseUint8Bytes(campos[0]),
		HighBP:               parseUint8Bytes(campos[1]),
		HighChol:             parseUint8Bytes(campos[2]),
		CholCheck:            parseUint8Bytes(campos[3]),
		BMI:                  parseUint8Bytes(campos[4]),
		Smoker:               parseUint8Bytes(campos[5]),
		Stroke:               parseUint8Bytes(campos[6]),
		HeartDiseaseorAttack: parseUint8Bytes(campos[7]),
		PhysActivity:         parseUint8Bytes(campos[8]),
		Fruits:               parseUint8Bytes(campos[9]),
		Veggies:              parseUint8Bytes(campos[10]),
		HvyAlcoholConsump:    parseUint8Bytes(campos[11]),
		AnyHealthcare:        parseUint8Bytes(campos[12]),
		NoDocbcCost:          parseUint8Bytes(campos[13]),
		GenHlth:              parseUint8Bytes(campos[14]),
		MentHlth:             parseFloatBytes(campos[15]),
		PhysHlth:             parseFloatBytes(campos[16]),
		DiffWalk:             parseUint8Bytes(campos[17]),
		Sex:                  parseUint8Bytes(campos[18]),
		Age:                  parseUint8Bytes(campos[19]),
		Education:            parseUint8Bytes(campos[20]),
		Income:               parseUint8Bytes(campos[21]),
	}
}

func parseUint8Bytes(valor []byte) uint8 {
	numero, _ := strconv.ParseUint(string(valor), 10, 8)
	return uint8(numero)
}

func parseFloatBytes(valor []byte) float64 {
	numero, _ := strconv.ParseFloat(string(valor), 64)
	return numero
}

func particionarDatos(data []models.PerfilPaciente, size int) [][]models.PerfilPaciente {
	chunks := make([][]models.PerfilPaciente, 0, (len(data)+size-1)/size)
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
