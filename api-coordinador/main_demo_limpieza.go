//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"api-coordinador/internal/limpieza"
	"api-coordinador/internal/loader"
)

func main() {
	inicio := time.Now()

	numWorkers := 12
	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		datasetPath = "../datos_raw/diabetes_1M_extended.csv"
	}

	fmt.Println("==========================================================================")
	fmt.Printf("Iniciando DEMOSTRACIÓN de Limpieza con %d Goroutines\n", numWorkers)
	fmt.Printf("Dataset: %s\n", datasetPath)
	fmt.Println("==========================================================================")

	jobs := make(chan []string, 10000)
	go loader.LeerCSVMasivo(datasetPath, jobs)

	canalLimpio := limpieza.IniciarWorkerPoolCompacto(numWorkers, jobs)

	count := 0
	for bytesLimpios := range canalLimpio {
		count++

		if count <= 5 {
			fmt.Printf("\n--- Paciente #%d Procesado por Limpieza ---\n", count)
			fmt.Printf("Resultado en Bytes (Listo para TCP): %v\n", bytesLimpios[:30])
			fmt.Printf("Resultado Decodificado (Parseo Seguro): %s\n", string(bytesLimpios))
			fmt.Println("Nota: Si había un campo vacío, parseWithDefault() inyectó un '0' automáticamente.")
		}
	}

	fmt.Println("\n==========================================================================")
	fmt.Printf("Limpieza Completada con Éxito\n")
	fmt.Printf("Total de Pacientes Procesados: %d\n", count)
	fmt.Printf("Tiempo Total (Lectura + Parseo Concurrente): %s\n", time.Since(inicio))
	fmt.Println("==========================================================================")
}
