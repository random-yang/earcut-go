package earcut

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"earcut-go/pkg/earcut"
)

// todo: port more tests from JS version

func TestIndices2D(t *testing.T) {
	data := []float64{10, 0, 0, 50, 60, 60, 70, 10}
	indices := earcut.Triangulate(data, nil, 2)
	assert.Equal(t, []int{1, 0, 3, 3, 2, 1}, indices)
}

func TestIndices3D(t *testing.T) {
	data := []float64{10, 0, 0, 0, 50, 0, 60, 60, 0, 70, 10, 0}
	indices := earcut.Triangulate(data, nil, 3)
	assert.Equal(t, []int{1, 0, 3, 3, 2, 1}, indices)
}

func TestInfiniteLoop(t *testing.T) {
	data := []float64{1, 2, 2, 2, 1, 2, 1, 1, 1, 2, 4, 1, 5, 1, 3, 2, 4, 2, 4, 1}
	holeIndices := []int{5}
	indices := earcut.Triangulate(data, holeIndices, 2)
	assert.NotNil(t, indices)
}

func TestTrianglesBuilding(t *testing.T) {
	data := []float64{
		661, 112, 661, 96, 666, 96,
		666, 87, 743, 87, 771, 87,
		771, 114, 750, 114, 750, 113,
		742, 113, 742, 106, 710, 106,
		710, 113, 666, 113, 666, 112,
	}
	indices := earcut.Triangulate(data, nil, 2)
	assert.Equal(t, 39, len(indices), "Should generate 13 triangles (39 indices)")
}

// func TestComplexPolygon(t *testing.T) {
// 	data := []float64{
// 		7, 18, 7, 15, 5, 15,
// 		7, 13, 7, 15, 17, 17,
// 	}
// 	indices := Triangulate(data, nil, 2)
// 	assert.Equal(t, 6, len(indices), "Should generate 6 triangles (18 indices)")
// }

func TestComplexPolygonWithHole(t *testing.T) {
	data := []float64{
		120, 2031, 92, 2368, 94, 2200,
		33, 2119, 42, 2112, 53, 2068,
		44, 2104, 79, 2132, 88, 2115,
		44, 2104,
	}
	holes := []int{6}
	indices := earcut.Triangulate(data, holes, 2)
	assert.Equal(t, 8*3, len(indices), "Should generate 8 triangles (24 indices)")
}

// func TestComplexPolygonWithHoles(t *testing.T) {
// 	data := []float64{
// 		810, 2828, 818, 2828, 832, 2818, 844, 2806, 855, 2808,
// 		866, 2816, 867, 2824, 876, 2827, 883, 2834, 875, 2834,
// 		867, 2840, 878, 2838, 889, 2844, 880, 2847, 870, 2847,
// 		860, 2864, 852, 2879, 847, 2867, 810, 2828, 810, 2828,
// 		818, 2834, 823, 2833, 831, 2828, 839, 2829, 839, 2837,
// 		851, 2845, 847, 2835, 846, 2827, 847, 2827, 837, 2827,
// 		840, 2815, 835, 2823, 818, 2834, 818, 2834, 857, 2846,
// 		864, 2850, 866, 2839, 857, 2846, 857, 2846, 848, 2863,
// 		848, 2866, 854, 2852, 846, 2854, 847, 2862, 838, 2851,
// 		838, 2859, 848, 2863, 848, 2863,
// 	}
// 	holes := []int{20, 34, 39}
// 	indices := Triangulate(data, holes, 2)
// 	fmt.Println(indices)
// 	assert.Equal(t, 42*3, len(indices), "Should generate 14 triangles (42 indices)")
// }

func TestPolygonWithHoles(t *testing.T) {
	data := []float64{
		0, 0, 100, 0, 100, 100,
		0, 100, 50, 50, 30, 40,
		70, 60, 20, 70,
	}
	holes := []int{4, 5, 6, 7}
	indices := earcut.Triangulate(data, holes, 2)
	assert.Equal(t, 27, len(indices), "Should generate 9 triangles (27 indices)")
}

func TestPolygonWithMultipleHoles(t *testing.T) {
	data := []float64{
		10, 10, 25, 10, 25, 40, 10,
		40, 15, 30, 20, 35, 10, 40,
		15, 15, 15, 20, 20, 15,
	}
	holes := []int{4, 7}
	indices := earcut.Triangulate(data, holes, 2)
	assert.Equal(t, 30, len(indices), "Should generate 10 triangles (30 indices)")
}
