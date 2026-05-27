# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependencies dùng chung từ root backend
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg (Models, DB logic dùng chung)
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Software Service
COPY service/software/ ./service/software/

# Build binary từ thư mục cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o software-service ./service/software/cmd/

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/software-service .

# Port khớp với cấu hình hệ thống
EXPOSE 8012

CMD ["./software-service"]