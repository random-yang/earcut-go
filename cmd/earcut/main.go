package main

import (
	"fmt"

	"earcut-go/pkg/earcut"
)

func main() {
	// 示例：一个简单的正方形
	vertices := []float64{
		0, 0,
		1, 0,
		1, 1,
		0, 1,
	}

	// 进行三角剖分
	triangles := earcut.Triangulate(vertices, nil, 2)

	fmt.Println("三角形索引：")
	for i := 0; i < len(triangles); i += 3 {
		fmt.Printf("Triangle %d: [%d, %d, %d]\n", i/3, triangles[i], triangles[i+1], triangles[i+2])
	}
}
