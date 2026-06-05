package engine

import (
	"nodo-ml/internal/models"
)

type LocalStats struct {
	ConteoClase     [3]int
	FreqCategoricas [3][21][100]int
	SumaContinuas   [3][3]float64
	SumaSqContinuas [3][3]float64
}

type NaiveBayesModel struct {
	ProbPriori        [3]float64
	ProbCategoricas   [3][21][100]float64
	MediaContinuas    [3][3]float64
	VarianzaContinuas [3][3]float64
}

// Ejecuta la clasificación paramétrica
func EntrenarNaiveBayes(chunks [][]models.PerfilPaciente) NaiveBayesModel {
	chStats := make(chan LocalStats, len(chunks))

	// Workers
	for _, chunk := range chunks {
		go func(data []models.PerfilPaciente) {
			var local LocalStats
			for _, p := range data {
				c := p.Diabetes012
				local.ConteoClase[c]++

				local.SumaContinuas[c][0] += float64(p.BMI)
				local.SumaSqContinuas[c][0] += float64(p.BMI) * float64(p.BMI)

				local.SumaContinuas[c][1] += p.MentHlth
				local.SumaSqContinuas[c][1] += p.MentHlth * p.MentHlth

				local.SumaContinuas[c][2] += p.PhysHlth
				local.SumaSqContinuas[c][2] += p.PhysHlth * p.PhysHlth

				features := extraerFeatures(&p)
				for f := 0; f < 21; f++ {
					if f == 3 || f == 14 || f == 15 {
						continue
					}
					val := int(features[f])
					if val < 100 {
						local.FreqCategoricas[c][f][val]++
					}
				}
			}
			chStats <- local
		}(chunk)
	}

	var global LocalStats
	for i := 0; i < len(chunks); i++ {
		local := <-chStats
		for c := 0; c < 3; c++ {
			global.ConteoClase[c] += local.ConteoClase[c]
			for k := 0; k < 3; k++ {
				global.SumaContinuas[c][k] += local.SumaContinuas[c][k]
				global.SumaSqContinuas[c][k] += local.SumaSqContinuas[c][k]
			}
			for f := 0; f < 21; f++ {
				for v := 0; v < 100; v++ {
					global.FreqCategoricas[c][f][v] += local.FreqCategoricas[c][f][v]
				}
			}
		}
	}

	return construirModelo(global)
}

// Aplica suavizado de Laplace y calcula parámetros paramétricos
func construirModelo(global LocalStats) NaiveBayesModel {
	var modelo NaiveBayesModel
	totalMuestras := global.ConteoClase[0] + global.ConteoClase[1] + global.ConteoClase[2]

	for c := 0; c < 3; c++ {
		nClase := float64(global.ConteoClase[c])
		if nClase == 0 {
			continue
		}

		modelo.ProbPriori[c] = nClase / float64(totalMuestras)

		// Parámetros de Distribución Gaussiana
		for k := 0; k < 3; k++ {
			media := global.SumaContinuas[c][k] / nClase
			varianza := (global.SumaSqContinuas[c][k] / nClase) - (media * media)
			modelo.MediaContinuas[c][k] = media
			modelo.VarianzaContinuas[c][k] = varianza
		}

		// Distribución Multinomial + Suavizado Laplace
		for f := 0; f < 21; f++ {
			// Corrección: El denominador debe usar la cardinalidad real del arreglo iterado (100) para asegurar masa probabilística total = 1.0
			denominador := nClase + 100.0
			for v := 0; v < 100; v++ {
				modelo.ProbCategoricas[c][f][v] = (float64(global.FreqCategoricas[c][f][v]) + 1.0) / denominador
			}
		}
	}
	return modelo
}
