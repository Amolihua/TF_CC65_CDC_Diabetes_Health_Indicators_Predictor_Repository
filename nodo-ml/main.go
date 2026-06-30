package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"nodo-ml/internal/engine"
	"nodo-ml/internal/models"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type MetadataEntrenamiento struct {
	Algoritmo      string `json:"algoritmo"`
	NumWorkers     int    `json:"num_workers"`
	NumTrees       int    `json:"num_trees,omitempty"`
	TotalRegistros int    `json:"total_registros,omitempty"`
}

func main() {
	listener, _ := net.Listen("tcp", ":9000")
	fmt.Println("Nodo ML TCP Server escuchando en puerto: |-| 9000 |-| ")

	for {
		conn, _ := listener.Accept()
		go manejarConexion(conn)
	}
}

func manejarConexion(conn net.Conn) {
	defer conn.Close()

	inicio := time.Now()
	meta, dataset := recibirDataset(conn)
	numWorkers := resolverWorkers(meta.NumWorkers)

	fmt.Printf("[TCP] Recibida solicitud para motor: %s con %d registros | workers=%d | arboles_esperados=%d\n", meta.Algoritmo, len(dataset), numWorkers, meta.NumTrees)

	inicioEntrenamiento := time.Now()

	var respuestaBinaria []byte
	if meta.Algoritmo == "random_forest" {
		numTrees := meta.NumTrees
		if numTrees == 0 {
			numTrees = 10
		}
		bosque := engine.EntrenarRandomForest(dataset, numTrees, numWorkers)

		for _, arbol := range bosque {
			bytesArbol := engine.SerializeTree(arbol)
			respuestaBinaria = append(respuestaBinaria, bytesArbol...)
		}
	} else {
		fmt.Printf("[WARN] Algoritmo no soportado para clúster: %s\n", meta.Algoritmo)
	}

	tiempoEntrenamiento := time.Since(inicioEntrenamiento)

	// Enviar respuesta binaria directa
	_, _ = conn.Write(respuestaBinaria)

	fmt.Printf("[TCP] Proceso finalizado. Registros: %d | Árboles generados: %d | Total: %s | Entrenamiento: %s\n", len(dataset), meta.NumTrees, time.Since(inicio), tiempoEntrenamiento)
}

func recibirDataset(conn net.Conn) (MetadataEntrenamiento, []models.PerfilPaciente) {
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
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for raw := range rawJobs {
				parsedResults <- parsearPerfilCompacto(raw)
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

		clon := make([]byte, len(linea))
		copy(clon, linea)
		rawJobs <- clon
	}

	close(rawJobs)
	wg.Wait()
	close(parsedResults)
	<-done

	return meta, dataset
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
