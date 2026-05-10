# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cài đặt các công cụ cần thiết
RUN apk add --no-cache git

# Copy dependencies từ thư mục gốc backend (Vì dùng chung go.mod)
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg dùng chung (Models, Database, Utils, Auth logic)
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của dịch vụ Auth
COPY service/auth/ ./service/auth/

# Build file thực thi từ thư mục cmd của auth
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service service/auth/cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

# Copy binary từ builder
COPY --from=builder /app/auth-service .

# Port của Auth Service
EXPOSE 8001

CMD ["./auth-service"]