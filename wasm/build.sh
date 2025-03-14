#!/bin/bash

# 设置 GOOS 和 GOARCH 环境变量，以便编译为 WebAssembly
export GOOS=js
export GOARCH=wasm

# 编译 Go 代码为 WebAssembly
echo "编译 Go 代码为 WebAssembly..."
go build -o main.wasm main.go

# 复制 wasm_exec.js 文件（Go 提供的 JavaScript 胶水代码）
echo "复制 wasm_exec.js 文件..."
WASM_EXEC_JS="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC_JS" ]; then
    # 尝试旧路径
    WASM_EXEC_JS="$(go env GOROOT)/misc/wasm/wasm_exec.js"
    if [ ! -f "$WASM_EXEC_JS" ]; then
        # 尝试查找文件
        WASM_EXEC_JS=$(find "$(go env GOROOT)" -name "wasm_exec.js" | head -n 1)
        if [ -z "$WASM_EXEC_JS" ]; then
            echo "错误：无法找到 wasm_exec.js 文件"
            exit 1
        fi
    fi
fi

echo "使用 wasm_exec.js 文件：$WASM_EXEC_JS"
cp "$WASM_EXEC_JS" .

echo "编译完成！"
echo "现在您可以使用 HTTP 服务器来提供这些文件："
echo "  - main.wasm"
echo "  - wasm_exec.js"
echo "  - index.html"
echo ""
echo "例如，您可以使用以下命令启动一个简单的 HTTP 服务器："
echo "  python3 -m http.server"
echo "然后在浏览器中访问 http://localhost:8000" 