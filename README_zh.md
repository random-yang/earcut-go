# Earcut-Go

> 声明：这是一个玩具项目，初衷是为了测试编程代理的能力。

这是一个用 Go 语言实现的多边形三角剖分库，基于 [earcut](https://github.com/mapbox/earcut) 算法。

## 功能特性

- 支持处理简单多边形
- 支持处理带洞的多边形
- 高效的三角剖分算法
- 完全用 Go 语言实现，无外部依赖
- 支持编译为 WebAssembly 并在浏览器中使用

## 安装

```bash
go get github.com/yourusername/earcut-go
```

## 使用示例

### Go 语言使用

```go
package main

import (
    "fmt"
    "github.com/yourusername/earcut-go/pkg/earcut"
)

func main() {
    // 定义多边形顶点
    vertices := []float64{
        0, 0,  // 第一个顶点
        1, 0,  // 第二个顶点
        1, 1,  // 第三个顶点
        0, 1,  // 第四个顶点
    }

    // 进行三角剖分
    triangles := earcut.Triangulate(vertices, nil, 2)
    fmt.Println(triangles)
}
```

### WebAssembly 使用

本库支持编译为 WebAssembly 并在浏览器中使用。详细说明请参考 [wasm/README.md](wasm/README_zh.md)。

简单示例：

```javascript
// 加载 WASM
const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
    .then((result) => {
        go.run(result.instance);
        
        // 定义多边形顶点
        const vertices = [0, 0, 1, 0, 1, 1, 0, 1];
        
        // 进行三角剖分
        const triangles = earcutGo(vertices, [], 2);
        console.log(triangles);
    });
```

## 文档

详细的 API 文档请参考 [GoDoc](https://pkg.go.dev/github.com/yourusername/earcut-go)。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。