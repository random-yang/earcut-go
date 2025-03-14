#!/bin/bash

# Set GOOS and GOARCH environment variables for WebAssembly compilation
export GOOS=js
export GOARCH=wasm

# Compile Go code to WebAssembly
echo "Compiling Go code to WebAssembly..."
go build -o main.wasm main.go

# Copy wasm_exec.js file (JavaScript glue code provided by Go)
echo "Copying wasm_exec.js file..."
WASM_EXEC_JS="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC_JS" ]; then
    # Try old path
    WASM_EXEC_JS="$(go env GOROOT)/misc/wasm/wasm_exec.js"
    if [ ! -f "$WASM_EXEC_JS" ]; then
        # Try to find the file
        WASM_EXEC_JS=$(find "$(go env GOROOT)" -name "wasm_exec.js" | head -n 1)
        if [ -z "$WASM_EXEC_JS" ]; then
            echo "Error: Cannot find wasm_exec.js file"
            exit 1
        fi
    fi
fi

echo "Using wasm_exec.js file: $WASM_EXEC_JS"
cp "$WASM_EXEC_JS" .

echo "Compilation complete!"
echo "Now you can use an HTTP server to serve these files:"
echo "  - main.wasm"
echo "  - wasm_exec.js"
echo "  - index.html"
echo ""
echo "For example, you can start a simple HTTP server with the following command:"
echo "  python3 -m http.server"
echo "Then visit http://localhost:8000 in your browser" 