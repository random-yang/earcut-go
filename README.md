# Earcut-Go

> Note: This is a toy project, initially created to test the capabilities of programming agents.

This is a polygon triangulation library implemented in Go, based on the [earcut](https://github.com/mapbox/earcut) algorithm.

## Features

- Support for simple polygons
- Support for polygons with holes
- Efficient triangulation algorithm
- Fully implemented in Go with no external dependencies
- Support for compilation to WebAssembly for use in browsers

## Documentation

[English](README.md) | [中文文档](README_zh.md)

## Installation

```bash
go get github.com/yourusername/earcut-go
```

## Usage Examples

### Go Usage

```go
package main

import (
    "fmt"
    "github.com/yourusername/earcut-go/pkg/earcut"
)

func main() {
    // Define polygon vertices
    vertices := []float64{
        0, 0,  // First vertex
        1, 0,  // Second vertex
        1, 1,  // Third vertex
        0, 1,  // Fourth vertex
    }

    // Perform triangulation
    triangles := earcut.Triangulate(vertices, nil, 2)
    fmt.Println(triangles)
}
```

### WebAssembly Usage

This library supports compilation to WebAssembly for use in browsers. For detailed instructions, please refer to [wasm/README.md](wasm/README.md).

Simple example:

```javascript
// Load WASM
const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
    .then((result) => {
        go.run(result.instance);
        
        // Define polygon vertices
        const vertices = [0, 0, 1, 0, 1, 1, 0, 1];
        
        // Perform triangulation
        const triangles = earcutGo(vertices, [], 2);
        console.log(triangles);
    });
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details. 