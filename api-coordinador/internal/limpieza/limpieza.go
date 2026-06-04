package limpieza

import (
	"api-coordinador/internal/models"
	"strconv"
	"sync"
)

func IniciarWorkerPool(numWorkers int, jobs <-chan []string) <-chan models.PerfilPaciente {
	results := make(chan models.PerfilPaciente, 10000)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go workerParser(&wg, jobs, results)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func workerParser(wg *sync.WaitGroup, jobs <-chan []string, results chan<- models.PerfilPaciente) {
	defer wg.Done()

	for record := range jobs {
		perfil := models.PerfilPaciente{
			Diabetes012:          parseWithDefaultUint8(record[0], 0),
			HighBP:               parseWithDefaultUint8(record[1], 0),
			HighChol:             parseWithDefaultUint8(record[2], 0),
			CholCheck:            parseWithDefaultUint8(record[3], 1),
			BMI:                  parseWithDefaultUint8(record[4], 27),
			Smoker:               parseWithDefaultUint8(record[5], 0),
			Stroke:               parseWithDefaultUint8(record[6], 0),
			HeartDiseaseorAttack: parseWithDefaultUint8(record[7], 0),
			PhysActivity:         parseWithDefaultUint8(record[8], 1),
			Fruits:               parseWithDefaultUint8(record[9], 1),
			Veggies:              parseWithDefaultUint8(record[10], 1),
			HvyAlcoholConsump:    parseWithDefaultUint8(record[11], 0),
			AnyHealthcare:        parseWithDefaultUint8(record[12], 1),
			NoDocbcCost:          parseWithDefaultUint8(record[13], 0),
			GenHlth:              parseWithDefaultUint8(record[14], 2),
			MentHlth:             parseWithDefaultFloat64(record[15], 0.0),
			PhysHlth:             parseWithDefaultFloat64(record[16], 0.0),
			DiffWalk:             parseWithDefaultUint8(record[17], 0),
			Sex:                  parseUint8(record[18]),
			Age:                  parseWithDefaultUint8(record[19], 9),
			Education:            parseWithDefaultUint8(record[20], 5),
			Income:               parseWithDefaultUint8(record[21], 6),
		}
		results <- perfil
	}
}
func parseWithDefaultUint8(s string, defaultVal uint8) uint8 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return uint8(v)
}

func parseWithDefaultFloat64(s string, defaultVal float64) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseUint8(s string) uint8 {
	v, _ := strconv.ParseFloat(s, 64)
	return uint8(v)
}
