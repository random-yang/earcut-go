package earcut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndices2D(t *testing.T) {
	data := []float64{10, 0, 0, 50, 60, 60, 70, 10}
	indices := Triangulate(data, nil, 2)
	assert.Equal(t, []int{1, 0, 3, 3, 2, 1}, indices)
}

func TestIndices3D(t *testing.T) {
	data := []float64{10, 0, 0, 0, 50, 0, 60, 60, 0, 70, 10, 0}
	indices := Triangulate(data, nil, 3)
	assert.Equal(t, []int{1, 0, 3, 3, 2, 1}, indices)
}
