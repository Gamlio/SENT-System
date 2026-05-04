#!/bin/bash
# build.sh

APP_NAME="SENT"
VERSION="4.0"
OUTPUT_DIR="dist"

# Tạo thư mục phân phối
mkdir -p $OUTPUT_DIR/data
mkdir -p $OUTPUT_DIR/config

echo "🚀 Biên dịch $APP_NAME v$VERSION..."

# Build cho Windows (Ẩn cửa sổ console với -ldflags="-s -w -H=windowsgui" nếu cần)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${APP_NAME}_windows.exe cmd/main.go

# Build cho Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${APP_NAME}_linux cmd/main.go

# Copy file .env mẫu nếu có
cp .env.example $OUTPUT_DIR/config/.env 2>/dev/null

echo "✅ Đã xong! Sản phẩm nằm trong thư mục /$OUTPUT_DIR"