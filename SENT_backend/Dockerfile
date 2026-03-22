# Bước 1: Build binary
FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go

# Bước 2: Chạy binary gọn nhẹ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
# Tạo thư mục cho uploads nếu có
RUN mkdir -p uploads

EXPOSE 8000
CMD ["./main"]