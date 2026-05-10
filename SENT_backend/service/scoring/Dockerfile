# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Tận dụng cache cho dependencies từ thư mục gốc
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg dùng chung (Models, DB logic)
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Scoring Service
COPY service/scoring/ ./service/scoring/

# Build binary từ thư mục cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o scoring-service ./service/scoring/cmd/

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/scoring-service .

# Port khớp với cấu hình hệ thống
EXPOSE 8010

CMD ["./scoring-service"]