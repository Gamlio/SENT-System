#!/bin/bash

APP_NAME="SENT"
VERSION="4.0"

echo "🚀 Đang bắt đầu biên dịch $APP_NAME v$VERSION..."

# 1. Build cho Windows (Bản chính cho máy trạm)
echo "📦 Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_windows.exe main.go

# 2. Build cho Linux (Dành cho Server hoặc máy trạm Linux)
echo "📦 Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_linux main.go

# 3. Build cho MacOS
echo "📦 Building for Darwin (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_mac main.go

echo "✅ Đã xong! Các file thực thi nằm trong thư mục /bin"