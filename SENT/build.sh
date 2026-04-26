#!/bin/bash
# build.sh

APP_NAME="SENT"
VERSION="4.0"

# Tạo thư mục bin nếu chưa có
mkdir -p bin

echo "🚀 Biên dịch $APP_NAME v$VERSION..."

# Build cho Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_windows.exe main.go

# Build cho Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_linux main.go

# Build cho MacOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/${APP_NAME}_mac main.go

echo "✅ Đã xong!"