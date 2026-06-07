package engine

import (
	"math/rand"
	"sync"
	"time"

	"nodo-ml/internal/models"
)

type TreeNode struct {
	FeatureIndex int
	Threshold    float64
	Left         *TreeNode
	Right        *TreeNode
	Value        uint8
	IsLeaf       bool
}

// Construye N árboles de decisión concurrentemente, limitados por numWorkers
func EntrenarRandomForest(data []models.PerfilPaciente, numTrees int, numWorkers int) []*TreeNode {
	var wg sync.WaitGroup
	bosque := make([]*TreeNode, numTrees)

	dataPtrs := make([]*models.PerfilPaciente, len(data))
	for i := range data {
		dataPtrs[i] = &data[i]
	}

	// Semáforo síncrono para limitar las goroutines concurrentes activas
	sem := make(chan struct{}, numWorkers)

	// Worker Pool
	for i := 0; i < numTrees; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Bagging usando punteros
			sample := make([]*models.PerfilPaciente, len(dataPtrs))
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(index)))
			for j := 0; j < len(dataPtrs); j++ {
				sample[j] = dataPtrs[rng.Intn(len(dataPtrs))]
			}
			bosque[index] = buildTree(sample, 0, 20, 50, rng)
		}(i)
	}

	wg.Wait()
	return bosque
}

// Construye el árbol recursivamente usando Punteros para evitar OOM
func buildTree(data []*models.PerfilPaciente, depth, maxDepth, minSamples int, rng *rand.Rand) *TreeNode {
	if depth >= maxDepth || len(data) < minSamples {
		return &TreeNode{IsLeaf: true, Value: majorityClass(data)}
	}

	// Selecciona un feature aleatorio
	featureIndex := rng.Intn(21)

	var sum float64
	for _, p := range data {
		val := ExtraerUnFeature(p, featureIndex)
		sum += val
	}
	threshold := sum / float64(len(data))
	i := 0
	j := len(data) - 1
	for i <= j {
		val := ExtraerUnFeature(data[i], featureIndex)
		if val <= threshold {
			i++
		} else {
			data[i], data[j] = data[j], data[i]
			j--
		}
	}
	leftData := data[:i]
	rightData := data[i:]

	if len(leftData) == 0 || len(rightData) == 0 {
		return &TreeNode{IsLeaf: true, Value: majorityClass(data)}
	}

	return &TreeNode{
		FeatureIndex: featureIndex,
		Threshold:    threshold,
		Left:         buildTree(leftData, depth+1, maxDepth, minSamples, rng),
		Right:        buildTree(rightData, depth+1, maxDepth, minSamples, rng),
		IsLeaf:       false,
	}
}

// Devuelve la clase ganadora aplicando Pesos de Clase (Class Weights)
func majorityClass(data []*models.PerfilPaciente) uint8 {
	var counts [3]int
	for _, p := range data {
		counts[p.Diabetes012]++
	}
	pesos := [3]float64{0.5, 30.0, 4.0}

	maxScore := -1.0
	var maxClass uint8
	for c, count := range counts {
		score := float64(count) * pesos[c]
		if score > maxScore {
			maxScore = score
			maxClass = uint8(c)
		}
	}
	return maxClass
}

// Optimización: Extraer solo el feature necesario, evitando asignar arreglos de 22 floats por cada nodo
func ExtraerUnFeature(p *models.PerfilPaciente, index int) float64 {
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
		return p.MentHlth
	case 15:
		return p.PhysHlth
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
	default:
		return 0.0
	}
}
