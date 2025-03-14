#!/bin/bash

# Check if Python is available
if command -v python3 &>/dev/null; then
    echo "Starting Python HTTP server..."
    python3 -m http.server
elif command -v python &>/dev/null; then
    echo "Starting Python HTTP server..."
    python -m SimpleHTTPServer
else
    echo "Error: Python not found. Please install Python or use another HTTP server."
    exit 1
fi 