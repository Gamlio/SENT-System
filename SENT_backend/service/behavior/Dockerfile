# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependencies từ root để đảm bảo đồng bộ module
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg (Models, DB, Middleware dùng chung)
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Behavior Service
# Đảm bảo chữ 'B' viết hoa đúng như tên thư mục trên ổ đĩa
COPY service/behavior/ ./service/behavior/

# Build binary từ thư mục cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o behavior-service ./service/behavior/cmd/main.go
# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/behavior-service .

# Port khớp với cấu hình hệ thống
EXPOSE 8003

CMD ["./behavior-service"]