#!/bin/bash

# 检查 Python 是否可用
if command -v python3 &>/dev/null; then
    echo "启动 Python HTTP 服务器..."
    python3 -m http.server
elif command -v python &>/dev/null; then
    echo "启动 Python HTTP 服务器..."
    python -m SimpleHTTPServer
else
    echo "错误：未找到 Python。请安装 Python 或使用其他 HTTP 服务器。"
    exit 1
fi 