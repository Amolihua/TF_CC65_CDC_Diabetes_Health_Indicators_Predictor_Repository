package evaluation

import (
	"fmt"
	"math"

	"nodo-ml/internal/engine"
	"nodo-ml/internal/models"
)

func PredecirSoftmax(p models.PerfilPaciente) uint8 {
	features := [22]float64{
		float64(p.HighBP), float64(p.HighChol), float64(p.CholCheck), float64(p.BMI),
		float64(p.Smoker), float64(p.Stroke), float64(p.HeartDiseaseorAttack), float64(p.PhysActivity),
		float64(p.Fruits), float64(p.Veggies), float64(p.HvyAlcoholConsump), float64(p.AnyHealthcare),
		float64(p.NoDocbcCost), float64(p.GenHlth), p.MentHlth, p.PhysHlth,
		float64(p.DiffWalk), float64(p.Sex), float64(p.Age), float64(p.Education), float64(p.Income),
		1.0,
	}

	maxLogit := -math.MaxFloat64
	var bestClass uint8

	for c := 0; c < 3; c++ {
		var z float64
		for j := 0; j < 22; j++ {
			z += engine.GlobalWeights[c][j] * features[j]
		}
		if z > maxLogit {
			maxLogit = z
			bestClass = uint8(c)
		}
	}
	return bestClass
}

func PredecirNaiveBayes(p models.PerfilPaciente, modelo engine.NaiveBayesModel) uint8 {
	features := [21]float64{
		float64(p.HighBP), float64(p.HighChol), float64(p.CholCheck), float64(p.BMI),
		float64(p.Smoker), float64(p.Stroke), float64(p.HeartDiseaseorAttack), float64(p.PhysActivity),
		float64(p.Fruits), float64(p.Veggies), float64(p.HvyAlcoholConsump), float64(p.AnyHealthcare),
		float64(p.NoDocbcCost), float64(p.GenHlth), p.MentHlth, p.PhysHlth,
		float64(p.DiffWalk), float64(p.Sex), float64(p.Age), float64(p.Education), float64(p.Income),
	}

	maxProb := -math.MaxFloat64
	var bestClass uint8

	for c := 0; c < 3; c++ {
		score := math.Log(modelo.ProbPriori[c])

		for k, fIdx := range []int{3, 14, 15} {
			val := features[fIdx]
			media := modelo.MediaContinuas[c][k]
			varianza := modelo.VarianzaContinuas[c][k]
			exponente := math.Exp(-math.Pow(val-media, 2) / (2 * varianza))
			densidad := (1.0 / math.Sqrt(2*math.Pi*varianza)) * exponente
			score += math.Log(densidad)
		}

		for f := 0; f < 21; f++ {
			if f == 3 || f == 14 || f == 15 {
				continue
			}
			val := int(features[f])
			score += math.Log(modelo.ProbCategoricas[c][f][val])
		}

		if score > maxProb {
			maxProb = score
			bestClass = uint8(c)
		}
	}
	return bestClass
}

func PredecirRandomForest(p models.PerfilPaciente, bosque []*engine.TreeNode) uint8 {
	var votos [3]int
	for _, arbol := range bosque {
		clase := predecirArbol(&p, arbol)
		votos[clase]++
	}

	maxVotos := -1
	mejorClase := uint8(0)
	for c, v := range votos {
		if v > maxVotos {
			maxVotos = v
			mejorClase = uint8(c)
		}
	}
	return mejorClase
}

func predecirArbol(p *models.PerfilPaciente, nodo *engine.TreeNode) uint8 {
	if nodo == nil {
		return 0
	}
	if nodo.IsLeaf {
		return nodo.Value
	}
	valorFeature := engine.ExtraerUnFeature(p, nodo.FeatureIndex)
	if valorFeature <= nodo.Threshold {
		return predecirArbol(p, nodo.Left)
	}
	return predecirArbol(p, nodo.Right)
}

func EjecutarCrossValidation(data []models.PerfilPaciente, algoritmo string) {
	var confusionMatrix [3][3]int
	foldSize := len(data) / 5

	for i := 0; i < 5; i++ {
		startTest := i * foldSize
		endTest := startTest + foldSize

		trainData := make([]models.PerfilPaciente, 0, len(data)-foldSize)
		trainData = append(trainData, data[:startTest]...)
		trainData = append(trainData, data[endTest:]...)
		testData := data[startTest:endTest]

		chunkSize := len(trainData) / 16
		chunks := make([][]models.PerfilPaciente, 0, 16)
		for j := 0; j < len(trainData); j += chunkSize {
			end := j + chunkSize
			if end > len(trainData) {
				end = len(trainData)
			}
			chunks = append(chunks, trainData[j:end])
		}

		if algoritmo == "softmax" {
			engine.GlobalWeights = [3][22]float64{}
			engine.EntrenarSoftmax(chunks, 0.005, 150)
			for _, p := range testData {
				pred := PredecirSoftmax(p)
				confusionMatrix[p.Diabetes012][pred]++
			}
		} else if algoritmo == "naive_bayes" {
			modelo := engine.EntrenarNaiveBayes(chunks)
			for _, p := range testData {
				pred := PredecirNaiveBayes(p, modelo)
				confusionMatrix[p.Diabetes012][pred]++
			}
		} else if algoritmo == "random_forest" {
			bosque := engine.EntrenarRandomForest(trainData, 50, 16)
			for _, p := range testData {
				pred := PredecirRandomForest(p, bosque)
				confusionMatrix[p.Diabetes012][pred]++
			}
		}
	}

	procesarYMostrarResultados(confusionMatrix)
}

// Calcula e imprime las métricas analíticas
func procesarYMostrarResultados(matriz [3][3]int) {
	fmt.Println("\n========================================================")
	fmt.Println("Matriz de Confusión [Fila=Real, Col=Predicha]:")
	fmt.Printf("   %7d %7d %7d\n", matriz[0][0], matriz[0][1], matriz[0][2])
	fmt.Printf("   %7d %7d %7d\n", matriz[1][0], matriz[1][1], matriz[1][2])
	fmt.Printf("   %7d %7d %7d\n\n", matriz[2][0], matriz[2][1], matriz[2][2])

	fmt.Println("Clase | Precision |  Recall  | F1-Score")
	fmt.Println("---------------------------------------")

	clases := []string{"0", "1", "2"}
	for c := 0; c < 3; c++ {
		tp := float64(matriz[c][c])
		fp := float64(matriz[(c+1)%3][c] + matriz[(c+2)%3][c])
		fn := float64(matriz[c][(c+1)%3] + matriz[c][(c+2)%3])

		precision := 0.0
		if tp+fp > 0 {
			precision = tp / (tp + fp)
		}
		recall := 0.0
		if tp+fn > 0 {
			recall = tp / (tp + fn)
		}
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}

		fmt.Printf("  %s   |  %7.4f  | %7.4f  | %7.4f\n", clases[c], precision, recall, f1)
	}
	fmt.Println("========================================================\n")
}
