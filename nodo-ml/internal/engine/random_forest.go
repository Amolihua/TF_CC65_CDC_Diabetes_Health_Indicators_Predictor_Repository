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

// Construye N árboles de decisión concurrentemente
func EntrenarRandomForest(data []models.PerfilPaciente, numTrees int) []*TreeNode {
	var wg sync.WaitGroup
	bosque := make([]*TreeNode, numTrees)

	// Worker Pool
	for i := 0; i < numTrees; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Bagging
			sample := make([]models.PerfilPaciente, len(data))
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(index)))
			for j := 0; j < len(data); j++ {
				sample[j] = data[rng.Intn(len(data))]
			}
			bosque[index] = buildTree(sample, 0, 10, 50, rng)
		}(i)
	}

	wg.Wait()
	return bosque
}

// Cnstruye el árbol recursivamente
func buildTree(data []models.PerfilPaciente, depth, maxDepth, minSamples int, rng *rand.Rand) *TreeNode {
	if depth >= maxDepth || len(data) < minSamples {
		return &TreeNode{IsLeaf: true, Value: majorityClass(data)}
	}

	// Selecciona un feature aleatorio
	featureIndex := rng.Intn(21)

	var sum float64
	for _, p := range data {
		features := extraerFeatures(&p)
		sum += features[featureIndex]
	}
	threshold := sum / float64(len(data))

	var leftData, rightData []models.PerfilPaciente
	for _, p := range data {
		features := extraerFeatures(&p)
		if features[featureIndex] <= threshold {
			leftData = append(leftData, p)
		} else {
			rightData = append(rightData, p)
		}
	}

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

// Devuelve la clase más frecuente
func majorityClass(data []models.PerfilPaciente) uint8 {
	var counts [3]int
	for _, p := range data {
		counts[p.Diabetes012]++
	}
	maxCount := -1
	var maxClass uint8
	for c, count := range counts {
		if count > maxCount {
			maxCount = count
			maxClass = uint8(c)
		}
	}
	return maxClass
}
