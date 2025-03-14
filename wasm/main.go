//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"earcut-go/pkg/earcut"
)

func main() {
	c := make(chan struct{}, 0)

	// Register earcut function to JavaScript
	js.Global().Set("earcutGo", js.FuncOf(earcutWrapper))

	// Keep the program running
	<-c
}

// earcutWrapper is a JavaScript wrapper for the earcut function
func earcutWrapper(this js.Value, args []js.Value) interface{} {
	// Check parameters
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "At least one parameter is required: vertex array",
		})
	}

	// Get vertex array
	jsData := args[0]
	dataLen := jsData.Length()
	data := make([]float64, dataLen)
	for i := 0; i < dataLen; i++ {
		data[i] = jsData.Index(i).Float()
	}

	// Get optional hole indices array
	var holeIndices []int
	if len(args) > 1 && !args[1].IsNull() && !args[1].IsUndefined() {
		jsHoles := args[1]
		holesLen := jsHoles.Length()
		holeIndices = make([]int, holesLen)
		for i := 0; i < holesLen; i++ {
			holeIndices[i] = jsHoles.Index(i).Int()
		}
	}

	// Get optional dimension parameter
	dim := 2
	if len(args) > 2 && !args[2].IsNull() && !args[2].IsUndefined() {
		dim = args[2].Int()
	}

	// Call earcut function
	triangles := earcut.Earcut(data, holeIndices, dim)

	// Convert result to JavaScript array
	jsTriangles := js.Global().Get("Array").New(len(triangles))
	for i, idx := range triangles {
		jsTriangles.SetIndex(i, idx)
	}

	return jsTriangles
}
