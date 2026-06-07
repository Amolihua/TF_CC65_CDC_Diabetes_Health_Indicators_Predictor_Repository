package limpieza

import (
	"api-coordinador/internal/models"
	"encoding/json"
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
		results <- construirPerfil(record)
	}
}

func IniciarWorkerPoolJSON(numWorkers int, jobs <-chan []string) <-chan []byte {
	results := make(chan []byte, 10000)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go workerParserJSON(&wg, jobs, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func workerParserJSON(wg *sync.WaitGroup, jobs <-chan []string, results chan<- []byte) {
	defer wg.Done()

	for record := range jobs {
		jsonBytes, _ := json.Marshal(construirPerfil(record))
		results <- jsonBytes
	}
}

func IniciarWorkerPoolCompacto(numWorkers int, jobs <-chan []string) <-chan []byte {
	results := make(chan []byte, 10000)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go workerParserCompacto(&wg, jobs, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func workerParserCompacto(wg *sync.WaitGroup, jobs <-chan []string, results chan<- []byte) {
	defer wg.Done()

	for record := range jobs {
		results <- serializarPerfilCompacto(construirPerfil(record))
	}
}

func construirPerfil(record []string) models.PerfilPaciente {
	return models.PerfilPaciente{
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
}

func serializarPerfilCompacto(p models.PerfilPaciente) []byte {
	buf := make([]byte, 0, 96)
	buf = appendUint8Field(buf, p.Diabetes012)
	buf = appendUint8Field(buf, p.HighBP)
	buf = appendUint8Field(buf, p.HighChol)
	buf = appendUint8Field(buf, p.CholCheck)
	buf = appendUint8Field(buf, p.BMI)
	buf = appendUint8Field(buf, p.Smoker)
	buf = appendUint8Field(buf, p.Stroke)
	buf = appendUint8Field(buf, p.HeartDiseaseorAttack)
	buf = appendUint8Field(buf, p.PhysActivity)
	buf = appendUint8Field(buf, p.Fruits)
	buf = appendUint8Field(buf, p.Veggies)
	buf = appendUint8Field(buf, p.HvyAlcoholConsump)
	buf = appendUint8Field(buf, p.AnyHealthcare)
	buf = appendUint8Field(buf, p.NoDocbcCost)
	buf = appendUint8Field(buf, p.GenHlth)
	buf = appendFloatField(buf, p.MentHlth)
	buf = appendFloatField(buf, p.PhysHlth)
	buf = appendUint8Field(buf, p.DiffWalk)
	buf = appendUint8Field(buf, p.Sex)
	buf = appendUint8Field(buf, p.Age)
	buf = appendUint8Field(buf, p.Education)
	return appendUint8Field(buf, p.Income)
}

func appendUint8Field(buf []byte, valor uint8) []byte {
	if len(buf) > 0 {
		buf = append(buf, ',')
	}
	return strconv.AppendUint(buf, uint64(valor), 10)
}

func appendFloatField(buf []byte, valor float64) []byte {
	if len(buf) > 0 {
		buf = append(buf, ',')
	}
	return strconv.AppendFloat(buf, valor, 'f', -1, 64)
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
