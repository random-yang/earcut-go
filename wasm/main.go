//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"earcut-go/pkg/earcut"
)

func main() {
	c := make(chan struct{}, 0)

	// 注册 earcut 函数到 JavaScript
	js.Global().Set("earcutGo", js.FuncOf(earcutWrapper))

	// 保持程序运行
	<-c
}

// earcutWrapper 是 earcut 函数的 JavaScript 包装器
func earcutWrapper(this js.Value, args []js.Value) interface{} {
	// 检查参数
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "需要至少一个参数：顶点数组",
		})
	}

	// 获取顶点数组
	jsData := args[0]
	dataLen := jsData.Length()
	data := make([]float64, dataLen)
	for i := 0; i < dataLen; i++ {
		data[i] = jsData.Index(i).Float()
	}

	// 获取可选的洞索引数组
	var holeIndices []int
	if len(args) > 1 && !args[1].IsNull() && !args[1].IsUndefined() {
		jsHoles := args[1]
		holesLen := jsHoles.Length()
		holeIndices = make([]int, holesLen)
		for i := 0; i < holesLen; i++ {
			holeIndices[i] = jsHoles.Index(i).Int()
		}
	}

	// 获取可选的维度参数
	dim := 2
	if len(args) > 2 && !args[2].IsNull() && !args[2].IsUndefined() {
		dim = args[2].Int()
	}

	// 调用 earcut 函数
	triangles := earcut.Earcut(data, holeIndices, dim)

	// 将结果转换为 JavaScript 数组
	jsTriangles := js.Global().Get("Array").New(len(triangles))
	for i, idx := range triangles {
		jsTriangles.SetIndex(i, idx)
	}

	return jsTriangles
}
