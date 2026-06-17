package models

import (
	"encoding/binary"
	"math"
)

type TreeNode struct {
	FeatureIndex int
	Threshold    float64
	Left         *TreeNode
	Right        *TreeNode
	Value        uint8
	IsLeaf       bool
}

// DeserializeTree reconstruye un árbol a partir del arreglo de bytes binario custom.
func DeserializeTree(data []byte) (*TreeNode, int) {
	if len(data) == 0 {
		return nil, 0
	}
	
	node := &TreeNode{}
	isLeafFlag := data[0]
	offset := 1
	
	if isLeafFlag == 1 {
		node.IsLeaf = true
		node.Value = data[offset]
		offset++
	} else {
		node.IsLeaf = false
		node.FeatureIndex = int(data[offset])
		offset++
		
		thresholdBits := binary.LittleEndian.Uint64(data[offset : offset+8])
		node.Threshold = math.Float64frombits(thresholdBits)
		offset += 8
		
		leftNode, leftBytesRead := DeserializeTree(data[offset:])
		node.Left = leftNode
		offset += leftBytesRead
		
		rightNode, rightBytesRead := DeserializeTree(data[offset:])
		node.Right = rightNode
		offset += rightBytesRead
	}
	
	return node, offset
}

// DeserializeForest toma un chorro completo de bytes y extrae los múltiples árboles.
func DeserializeForest(data []byte) []*TreeNode {
	var bosque []*TreeNode
	offset := 0
	for offset < len(data) {
		tree, bytesRead := DeserializeTree(data[offset:])
		if bytesRead == 0 {
			break
		}
		bosque = append(bosque, tree)
		offset += bytesRead
	}
	return bosque
}
