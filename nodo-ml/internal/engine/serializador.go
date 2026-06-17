package engine

import (
	"encoding/binary"
	"math"
)

func SerializeTree(node *TreeNode) []byte {
	if node == nil {
		return nil
	}
	buf := make([]byte, 0, 128)
	return serializeTreeRec(node, buf)
}

func serializeTreeRec(node *TreeNode, buf []byte) []byte {
	if node.IsLeaf {
		buf = append(buf, 1) // Flag: 1 = Hoja
		buf = append(buf, node.Value)
	} else {
		buf = append(buf, 0) // Flag: 0 = Nodo Interno
		buf = append(buf, byte(node.FeatureIndex))

		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], math.Float64bits(node.Threshold))
		buf = append(buf, b[:]...)

		// Recursión
		buf = serializeTreeRec(node.Left, buf)
		buf = serializeTreeRec(node.Right, buf)
	}
	return buf
}
