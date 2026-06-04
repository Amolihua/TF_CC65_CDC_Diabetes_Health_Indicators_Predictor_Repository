package analisis

import (
	"math"
	"sync"

	"api-coordinador/internal/models"
)


// Estadísticas locales mantenidas por cada worker
type estadisticasLocales struct {
	FreqDiabetes012 [3]int
	FreqHighBP      [2]int
	FreqHighChol    [2]int
	FreqSex         [2]int
	FreqMentHlth    [31]int
	FreqPhysHlth    [31]int
}

// Procesamiento de stream de datos limpios mediante un esquema Map-Reduce
func EjecutarAnalisisConcurrente(numWorkers int, dataStream <-chan models.PerfilPaciente) models.ReporteEstadistico {
	var wg sync.WaitGroup
	resultadosLocales := make(chan estadisticasLocales, numWorkers)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var local estadisticasLocales

			for perfil := range dataStream {
				local.FreqDiabetes012[perfil.Diabetes012]++
				local.FreqHighBP[perfil.HighBP]++
				local.FreqHighChol[perfil.HighChol]++
				local.FreqSex[perfil.Sex]++

				local.FreqMentHlth[int(perfil.MentHlth)]++
				local.FreqPhysHlth[int(perfil.PhysHlth)]++
			}

			resultadosLocales <- local
		}()
	}

	go func() {
		wg.Wait()
		close(resultadosLocales)
	}()

	var consolidado estadisticasLocales
	for local := range resultadosLocales {
		for i := 0; i < 3; i++ {
			consolidado.FreqDiabetes012[i] += local.FreqDiabetes012[i]
		}
		for i := 0; i < 2; i++ {
			consolidado.FreqHighBP[i] += local.FreqHighBP[i]
		}
		for i := 0; i < 2; i++ {
			consolidado.FreqHighChol[i] += local.FreqHighChol[i]
		}
		for i := 0; i < 2; i++ {
			consolidado.FreqSex[i] += local.FreqSex[i]
		}
		for i := 0; i < 31; i++ {
			consolidado.FreqMentHlth[i] += local.FreqMentHlth[i]
			consolidado.FreqPhysHlth[i] += local.FreqPhysHlth[i]
		}
	}

	return construirReporte(consolidado)
}

func construirReporte(c estadisticasLocales) models.ReporteEstadistico {
	return models.ReporteEstadistico{
		FreqDiabetes012: map[uint8]int{0: c.FreqDiabetes012[0], 1: c.FreqDiabetes012[1], 2: c.FreqDiabetes012[2]},
		FreqHighBP:      map[uint8]int{0: c.FreqHighBP[0], 1: c.FreqHighBP[1]},
		FreqHighChol:    map[uint8]int{0: c.FreqHighChol[0], 1: c.FreqHighChol[1]},
		FreqSex:         map[uint8]int{0: c.FreqSex[0], 1: c.FreqSex[1]},
		MentHlth:        calcularMetricas(c.FreqMentHlth),
		PhysHlth:        calcularMetricas(c.FreqPhysHlth),
	}
}

func calcularMetricas(freqs [31]int) models.MetricasNumericas {
	var conteo int
	var suma float64
	var sumaSq float64

	for val, freq := range freqs {
		if freq > 0 {
			conteo += freq
			suma += float64(val * freq)
			sumaSq += float64(val * val * freq)
		}
	}

	if conteo == 0 {
		return models.MetricasNumericas{}
	}

	media := suma / float64(conteo)
	varianza := (sumaSq / float64(conteo)) - (media * media)
	desvStd := math.Sqrt(varianza)

	mediana := 0.0
	acumulado := 0
	mitad := conteo / 2
	esPar := conteo%2 == 0

	for val, freq := range freqs {
		if freq == 0 {
			continue
		}

		acumulado += freq
		if acumulado > mitad {
			mediana = float64(val)
			break
		} else if acumulado == mitad && esPar {
			siguiente := val + 1
			for siguiente < 31 && freqs[siguiente] == 0 {
				siguiente++
			}
			mediana = float64(val+siguiente) / 2.0
			break
		}
	}

	return models.MetricasNumericas{
		Conteo:   conteo,
		Media:    media,
		Varianza: varianza,
		DesvStd:  desvStd,
		Mediana:  mediana,
	}
}
