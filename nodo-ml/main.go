package main

import (
	"encoding/json"
	"fmt"
	"net"

	"nodo-ml/internal/engine"
	"nodo-ml/internal/models"
)

type SolicitudEntrenamiento struct {
	Algoritmo string                  `json:"algoritmo"`
	Dataset   []models.PerfilPaciente `json:"dataset"`
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

	var solicitud SolicitudEntrenamiento
	json.NewDecoder(conn).Decode(&solicitud)

	fmt.Printf("[TCP] Recibida solicitud para motor: %s con %d registros\n", solicitud.Algoritmo, len(solicitud.Dataset))

	switch solicitud.Algoritmo {
	case "softmax":
		chunks := particionarDatos(solicitud.Dataset, 1000)
		engine.EntrenarSoftmax(chunks, 0.01)
		fmt.Println("[SOFTMAX] Entrenamiento iterativo concluido exitosamente.")

	case "random_forest":
		engine.EntrenarRandomForest(solicitud.Dataset, 10)
		fmt.Println("[RANDOM FOREST] Consolidación de 10 árboles concluida exitosamente.")

	case "naive_bayes":
		chunks := particionarDatos(solicitud.Dataset, 1000)
		engine.EntrenarNaiveBayes(chunks)
		fmt.Println("[NAIVE BAYES] Map-Reduce estadístico concluido exitosamente.")
	}
}

func particionarDatos(data []models.PerfilPaciente, size int) [][]models.PerfilPaciente {
	var chunks [][]models.PerfilPaciente
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
