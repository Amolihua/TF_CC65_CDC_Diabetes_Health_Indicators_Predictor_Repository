package analisis

import (
	"bytes"
	"fmt"

	"strconv"
	"sync"
	"time"

	"api-coordinador/internal/models"
)

func EvaluarBosqueDistribuido(testDataRaw [][]byte, bosque []*models.TreeNode, numWorkers int) {
	fmt.Printf("\n[EVALUACIÓN CENTRALIZADA] Iniciando evaluación sobre %d registros con bosque de %d árboles...\n", len(testDataRaw), len(bosque))
	inicio := time.Now()

	chunkSize := len(testDataRaw) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	chunks := make([][][]byte, 0, numWorkers)
	for i := 0; i < len(testDataRaw); i += chunkSize {
		end := i + chunkSize
		if end > len(testDataRaw) {
			end = len(testDataRaw)
		}
		chunks = append(chunks, testDataRaw[i:end])
	}

	var wg sync.WaitGroup
	chMatrices := make(chan [3][3]int, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(c [][]byte) {
			defer wg.Done()
			var localMatrix [3][3]int

			for _, rowBytes := range c {
				p := parsearPerfilCompacto(rowBytes)
				prediccion := int(PredecirRandomForest(p, bosque))
				claseReal := int(p.Diabetes012)
				localMatrix[claseReal][prediccion]++
			}
			chMatrices <- localMatrix
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(chMatrices)
	}()

	var globalMatrix [3][3]int
	for mat := range chMatrices {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				globalMatrix[r][c] += mat[r][c]
			}
		}
	}

	tiempo := time.Since(inicio)
	fmt.Printf("[EVALUACIÓN] Map-Reduce completado en %s\n", tiempo)
	procesarYMostrarResultados(globalMatrix)
}

func PredecirRandomForest(p models.PerfilPaciente, bosque []*models.TreeNode) uint8 {
	var votos [3]int
	for _, arbol := range bosque {
		pred := predecirArbol(p, arbol)
		votos[pred]++
	}

	maxVotos := -1
	var claseGanadora uint8
	for c, v := range votos {
		if v > maxVotos {
			maxVotos = v
			claseGanadora = uint8(c)
		}
	}
	return claseGanadora
}

func predecirArbol(p models.PerfilPaciente, nodo *models.TreeNode) uint8 {
	if nodo.IsLeaf {
		return nodo.Value
	}
	valorFeature := ExtraerUnFeature(p, nodo.FeatureIndex)
	if float64(valorFeature) <= nodo.Threshold {
		return predecirArbol(p, nodo.Left)
	}
	return predecirArbol(p, nodo.Right)
}

func ExtraerUnFeature(p models.PerfilPaciente, index int) float64 {
	switch index {
	case 0:
		return float64(p.HighBP)
	case 1:
		return float64(p.HighChol)
	case 2:
		return float64(p.CholCheck)
	case 3:
		return float64(p.BMI)
	case 4:
		return float64(p.Smoker)
	case 5:
		return float64(p.Stroke)
	case 6:
		return float64(p.HeartDiseaseorAttack)
	case 7:
		return float64(p.PhysActivity)
	case 8:
		return float64(p.Fruits)
	case 9:
		return float64(p.Veggies)
	case 10:
		return float64(p.HvyAlcoholConsump)
	case 11:
		return float64(p.AnyHealthcare)
	case 12:
		return float64(p.NoDocbcCost)
	case 13:
		return float64(p.GenHlth)
	case 14:
		return float64(p.MentHlth)
	case 15:
		return float64(p.PhysHlth)
	case 16:
		return float64(p.DiffWalk)
	case 17:
		return float64(p.Sex)
	case 18:
		return float64(p.Age)
	case 19:
		return float64(p.Education)
	case 20:
		return float64(p.Income)
	}
	return 0
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

func procesarYMostrarResultados(cm [3][3]int) {
	fmt.Println("\n---------------- MATRIZ DE CONFUSION ----------------")
	fmt.Printf("                   Pred_Sano(0)  Pred_PreDiab(1)  Pred_Diab(2)\n")
	fmt.Printf("Real_Sano(0)     : %-13d %-16d %-12d\n", cm[0][0], cm[0][1], cm[0][2])
	fmt.Printf("Real_PreDiab(1)  : %-13d %-16d %-12d\n", cm[1][0], cm[1][1], cm[1][2])
	fmt.Printf("Real_Diab(2)     : %-13d %-16d %-12d\n", cm[2][0], cm[2][1], cm[2][2])
	fmt.Println("-----------------------------------------------------")

	fmt.Println("\nTabla 6")
	fmt.Println("Métricas analíticas de clasificación del modelo Random Forest con 50 árboles")
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("%-11s %-12s %-9s %s\n", "Clase", "Precision", "Recall", "F1-Score")
	fmt.Println("---------------------------------------------------------")

	for i := 0; i < 3; i++ {
		tp := float64(cm[i][i])
		var fp, fn float64
		for j := 0; j < 3; j++ {
			if i != j {
				fp += float64(cm[j][i])
				fn += float64(cm[i][j])
			}
		}

		precision := 0.0
		if (tp + fp) > 0 {
			precision = tp / (tp + fp)
		}
		recall := 0.0
		if (tp + fn) > 0 {
			recall = tp / (tp + fn)
		}
		f1 := 0.0
		if (precision + recall) > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}

		fmt.Printf("Clase %d     %-12.4f %-9.4f %.4f\n", i, precision, recall, f1)
	}
	fmt.Println("---------------------------------------------------------")
	fmt.Println("=====================================================")
}
