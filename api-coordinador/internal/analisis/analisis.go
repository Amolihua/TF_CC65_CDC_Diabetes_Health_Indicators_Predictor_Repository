package analisis

import (
	"math"
	"sync"

	"api-coordinador/internal/models"
)

// Matriz de arreglos de tamaño fijo para evitar colisiones de memoria en Map-Reduce
type estadisticasLocales struct {
	FreqDiabetes012   [3]int
	FreqHighBP        [2]int
	FreqHighChol      [2]int
	FreqCholCheck     [2]int
	FreqBMI           [100]int
	FreqSmoker        [2]int
	FreqStroke        [2]int
	FreqHeartDisease  [2]int
	FreqPhysActivity  [2]int
	FreqFruits        [2]int
	FreqVeggies       [2]int
	FreqHvyAlcohol    [2]int
	FreqAnyHealthcare [2]int
	FreqNoDocbcCost   [2]int
	FreqGenHlth       [6]int
	FreqMentHlth      [31]int
	FreqPhysHlth      [31]int
	FreqDiffWalk      [2]int
	FreqSex           [2]int
	FreqAge           [14]int
	FreqEducation     [7]int
	FreqIncome        [9]int
}

// EjecutarAnalisisConcurrente procesa el stream de datos limpios mediante un esquema Map-Reduce
func EjecutarAnalisisConcurrente(numWorkers int, dataStream <-chan models.PerfilPaciente) models.ReporteEstadistico {
	var wg sync.WaitGroup
	resultadosLocales := make(chan estadisticasLocales, numWorkers)

	// Mapeo indexado lock-free de todas las variables
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local estadisticasLocales

			for perfil := range dataStream {
				local.FreqDiabetes012[perfil.Diabetes012]++
				local.FreqHighBP[perfil.HighBP]++
				local.FreqHighChol[perfil.HighChol]++
				local.FreqCholCheck[perfil.CholCheck]++
				local.FreqBMI[perfil.BMI]++
				local.FreqSmoker[perfil.Smoker]++
				local.FreqStroke[perfil.Stroke]++
				local.FreqHeartDisease[perfil.HeartDiseaseorAttack]++
				local.FreqPhysActivity[perfil.PhysActivity]++
				local.FreqFruits[perfil.Fruits]++
				local.FreqVeggies[perfil.Veggies]++
				local.FreqHvyAlcohol[perfil.HvyAlcoholConsump]++
				local.FreqAnyHealthcare[perfil.AnyHealthcare]++
				local.FreqNoDocbcCost[perfil.NoDocbcCost]++
				local.FreqGenHlth[perfil.GenHlth]++
				local.FreqMentHlth[int(perfil.MentHlth)]++
				local.FreqPhysHlth[int(perfil.PhysHlth)]++
				local.FreqDiffWalk[perfil.DiffWalk]++
				local.FreqSex[perfil.Sex]++
				local.FreqAge[perfil.Age]++
				local.FreqEducation[perfil.Education]++
				local.FreqIncome[perfil.Income]++
			}
			resultadosLocales <- local
		}()
	}

	go func() {
		wg.Wait()
		close(resultadosLocales)
	}()

	//  Reducción e interconexión de todas las frecuencias locales
	var con estadisticasLocales
	for local := range resultadosLocales {
		for i := 0; i < 3; i++ {
			con.FreqDiabetes012[i] += local.FreqDiabetes012[i]
		}
		for i := 0; i < 2; i++ {
			con.FreqHighBP[i] += local.FreqHighBP[i]
			con.FreqHighChol[i] += local.FreqHighChol[i]
			con.FreqCholCheck[i] += local.FreqCholCheck[i]
			con.FreqSmoker[i] += local.FreqSmoker[i]
			con.FreqStroke[i] += local.FreqStroke[i]
			con.FreqHeartDisease[i] += local.FreqHeartDisease[i]
			con.FreqPhysActivity[i] += local.FreqPhysActivity[i]
			con.FreqFruits[i] += local.FreqFruits[i]
			con.FreqVeggies[i] += local.FreqVeggies[i]
			con.FreqHvyAlcohol[i] += local.FreqHvyAlcohol[i]
			con.FreqAnyHealthcare[i] += local.FreqAnyHealthcare[i]
			con.FreqNoDocbcCost[i] += local.FreqNoDocbcCost[i]
			con.FreqDiffWalk[i] += local.FreqDiffWalk[i]
			con.FreqSex[i] += local.FreqSex[i]
		}
		for i := 0; i < 6; i++ {
			con.FreqGenHlth[i] += local.FreqGenHlth[i]
		}
		for i := 0; i < 14; i++ {
			con.FreqAge[i] += local.FreqAge[i]
		}
		for i := 0; i < 7; i++ {
			con.FreqEducation[i] += local.FreqEducation[i]
		}
		for i := 0; i < 9; i++ {
			con.FreqIncome[i] += local.FreqIncome[i]
		}
		for i := 0; i < 100; i++ {
			con.FreqBMI[i] += local.FreqBMI[i]
		}
		for i := 0; i < 31; i++ {
			con.FreqMentHlth[i] += local.FreqMentHlth[i]
			con.FreqPhysHlth[i] += local.FreqPhysHlth[i]
		}
	}

	return construirReporte(con)
}

// construirReporte formatea el objeto global a partir de las estadísticas reducidas
func construirReporte(c estadisticasLocales) models.ReporteEstadistico {
	return models.ReporteEstadistico{
		FreqDiabetes012:   generarMap(c.FreqDiabetes012[:]),
		FreqHighBP:        generarMap(c.FreqHighBP[:]),
		FreqHighChol:      generarMap(c.FreqHighChol[:]),
		FreqCholCheck:     generarMap(c.FreqCholCheck[:]),
		FreqSmoker:        generarMap(c.FreqSmoker[:]),
		FreqStroke:        generarMap(c.FreqStroke[:]),
		FreqHeartDisease:  generarMap(c.FreqHeartDisease[:]),
		FreqPhysActivity:  generarMap(c.FreqPhysActivity[:]),
		FreqFruits:        generarMap(c.FreqFruits[:]),
		FreqVeggies:       generarMap(c.FreqVeggies[:]),
		FreqHvyAlcohol:    generarMap(c.FreqHvyAlcohol[:]),
		FreqAnyHealthcare: generarMap(c.FreqAnyHealthcare[:]),
		FreqNoDocbcCost:   generarMap(c.FreqNoDocbcCost[:]),
		FreqDiffWalk:      generarMap(c.FreqDiffWalk[:]),
		FreqSex:           generarMap(c.FreqSex[:]),
		FreqGenHlth:       generarMap(c.FreqGenHlth[:]),
		FreqAge:           generarMap(c.FreqAge[:]),
		FreqEducation:     generarMap(c.FreqEducation[:]),
		FreqIncome:        generarMap(c.FreqIncome[:]),
		BMIDistribuido:    calcularMetricas(c.FreqBMI[:]),
		MentHlth:          calcularMetricas(c.FreqMentHlth[:]),
		PhysHlth:          calcularMetricas(c.FreqPhysHlth[:]),
	}
}

// generarMap convierte los arreglos fijos en mapas dinámicos ignorando conteos nulos
func generarMap(slice []int) map[uint8]int {
	res := make(map[uint8]int)
	for k, v := range slice {
		if v > 0 {
			res[uint8(k)] = v
		}
	}
	return res
}

// calcularMetricas procesa los slices extraídos obteniendo la mediana en espacio O(1)
func calcularMetricas(freqs []int) models.MetricasNumericas {
	var conteo int
	var suma, sumaSq float64

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
			for siguiente < len(freqs) && freqs[siguiente] == 0 {
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
