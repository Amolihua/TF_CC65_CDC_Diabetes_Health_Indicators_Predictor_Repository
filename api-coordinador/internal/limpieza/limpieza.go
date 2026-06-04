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
			Diabetes012:          parseUint8(record[0]),
			HighBP:               parseUint8(record[1]),
			HighChol:             parseUint8(record[2]),
			CholCheck:            parseUint8(record[3]),
			BMI:                  parseUint8(record[4]),
			Smoker:               parseUint8(record[5]),
			Stroke:               parseUint8(record[6]),
			HeartDiseaseorAttack: parseUint8(record[7]),
			PhysActivity:         parseUint8(record[8]),
			Fruits:               parseUint8(record[9]),
			Veggies:              parseUint8(record[10]),
			HvyAlcoholConsump:    parseUint8(record[11]),
			AnyHealthcare:        parseUint8(record[12]),
			NoDocbcCost:          parseUint8(record[13]),
			GenHlth:              parseUint8(record[14]),
			MentHlth:             parseFloat64(record[15]),
			PhysHlth:             parseFloat64(record[16]),
			DiffWalk:             parseUint8(record[17]),
			Sex:                  parseUint8(record[18]),
			Age:                  parseUint8(record[19]),
			Education:            parseUint8(record[20]),
			Income:               parseUint8(record[21]),
		}
		results <- perfil
	}
}

func parseUint8(s string) uint8 {
	v, _ := strconv.ParseFloat(s, 64)
	return uint8(v)
}

func parseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
