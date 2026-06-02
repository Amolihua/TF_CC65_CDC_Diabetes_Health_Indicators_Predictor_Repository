package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func workerMultiplicador(jobs <-chan []string, results chan<- [][]string, wg *sync.WaitGroup, factor int) {
	defer wg.Done()

	for fila := range jobs {
		var bloque [][]string
		for i := 0; i < factor; i++ {
			bloque = append(bloque, fila)
		}
		results <- bloque
	}
}

func main() {
	tiempoInicio := time.Now()
	fmt.Println("Iniciando multiplicación del dataset")

	in, _ := os.Open("../datos_raw/diabetes_012_health_indicators_BRFSS2015.csv")
	defer in.Close()
	lector := csv.NewReader(in)

	out, _ := os.Create("../datos_raw/diabetes_1M_extended.csv")
	defer out.Close()
	escritor := csv.NewWriter(out)

	cabecera, _ := lector.Read()
	escritor.Write(cabecera)

	jobs := make(chan []string, 5000)
	results := make(chan [][]string, 5000)

	var wgWorkers sync.WaitGroup
	var wgWriter sync.WaitGroup

	wgWriter.Add(1)
	go func() {
		defer wgWriter.Done()
		for bloque := range results {
			for _, fila := range bloque {
				escritor.Write(fila)
			}
		}
		escritor.Flush()
	}()

	factorMultiplicacion := 10
	numWorkers := 12

	for w := 1; w <= numWorkers; w++ {
		wgWorkers.Add(1)
		go workerMultiplicador(jobs, results, &wgWorkers, factorMultiplicacion)
	}

	for {
		fila, err := lector.Read()
		if err == io.EOF {
			break
		}
		jobs <- fila
	}
	close(jobs)

	wgWorkers.Wait()
	close(results)

	wgWriter.Wait()
	fmt.Printf("Multiplicación finalizada en %s\n", time.Since(tiempoInicio))
}
