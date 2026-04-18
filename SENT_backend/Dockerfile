# Stage 1: Build
FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
# Build binary tĩnh hoàn toàn
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Stage 2: Runner
FROM alpine:latest
# Cài đặt ca-certificates để Backend có thể gửi Email (SMTP) qua HTTPS
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Chỉ copy file thực thi, KHÔNG copy .env để bảo mật và linh hoạt
COPY --from=builder /app/main .
RUN mkdir -p uploads

EXPOSE 8000
CMD ["./main"]