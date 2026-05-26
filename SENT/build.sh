#!/bin/bash
# build.sh

APP_NAME="SENT"
VERSION="4.2.6"
OUTPUT_DIR="dist"

mkdir -p $OUTPUT_DIR/data
mkdir -p $OUTPUT_DIR/config

echo "🚀 Biên dịch $APP_NAME v$VERSION..."

GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${APP_NAME}_v${VERSION}_windows.exe ./cmd

GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${APP_NAME}_v${VERSION}_linux ./cmd

GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $OUTPUT_DIR/${APP_NAME}_v${VERSION}_mac ./cmd

cp .env.example $OUTPUT_DIR/config/.env 2>/dev/null

echo "✅ Đã xong! Sản phẩm phần mềm nằm trong thư mục /$OUTPUT_DIR"