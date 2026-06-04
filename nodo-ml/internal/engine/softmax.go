package engine

import (
	"math"
	"sync"

	"nodo-ml/internal/models"
)

var GlobalWeights [3][22]float64
var mu sync.Mutex

// Regresión Logística Multinomial concurrentemente
func EntrenarSoftmax(chunks [][]models.PerfilPaciente, learningRate float64) {
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		// Goroutines
		go func(data []models.PerfilPaciente) {
			defer wg.Done()

			// Gradiente acumulado local
			var localGradients [3][22]float64
			mu.Lock()
			currentWeights := GlobalWeights
			mu.Unlock()

			for _, p := range data {
				features := extraerFeatures(&p)
				claseReal := p.Diabetes012

				var logits [3]float64
				maxLogit := -math.MaxFloat64
				for c := 0; c < 3; c++ {
					var z float64
					for j := 0; j < 22; j++ {
						z += currentWeights[c][j] * features[j]
					}
					logits[c] = z
					if z > maxLogit {
						maxLogit = z
					}
				}

				var sumaExp float64
				var probs [3]float64
				for c := 0; c < 3; c++ {
					probs[c] = math.Exp(logits[c] - maxLogit)
					sumaExp += probs[c]
				}

				for c := 0; c < 3; c++ {
					probs[c] /= sumaExp
				}

				// Backpropagation
				for c := 0; c < 3; c++ {
					errorClase := probs[c]
					if c == int(claseReal) {
						errorClase -= 1.0
					}
					for j := 0; j < 22; j++ {
						localGradients[c][j] += errorClase * features[j]
					}
				}
			}

			mu.Lock()
			for c := 0; c < 3; c++ {
				for j := 0; j < 22; j++ {
					GlobalWeights[c][j] -= learningRate * localGradients[c][j]
				}
			}
			mu.Unlock()

		}(chunk)
	}

	wg.Wait()
}

func extraerFeatures(p *models.PerfilPaciente) [22]float64 {
	return [22]float64{
		float64(p.HighBP), float64(p.HighChol), float64(p.CholCheck), float64(p.BMI),
		float64(p.Smoker), float64(p.Stroke), float64(p.HeartDiseaseorAttack), float64(p.PhysActivity),
		float64(p.Fruits), float64(p.Veggies), float64(p.HvyAlcoholConsump), float64(p.AnyHealthcare),
		float64(p.NoDocbcCost), float64(p.GenHlth), p.MentHlth, p.PhysHlth,
		float64(p.DiffWalk), float64(p.Sex), float64(p.Age), float64(p.Education), float64(p.Income),
		1.0,
	}
}
