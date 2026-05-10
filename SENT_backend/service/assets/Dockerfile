# Stage 1: Build Stage
FROM golang:1.26-alpine AS builder

# Thiết lập thư mục làm việc
WORKDIR /app

# Cài đặt các công cụ cần thiết để build (nếu có dùng CGO hoặc thư viện hệ thống)
RUN apk add --no-cache git

# Copy file go.mod và go.sum từ thư mục gốc của SENT_backend để tận dụng cache layer
# Lưu ý: Khi build, chúng ta sẽ đứng ở thư mục SENT_backend làm context
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg (chứa models, database, middleware dùng chung)
COPY pkg/ ./pkg/

# Copy mã nguồn của riêng service assets
COPY service/assets/ ./service/assets/

# Build binary cho service assets
# Điểm bắt đầu là file main.go trong thư mục cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o asset-service service/assets/cmd/main.go

# Stage 2: Runtime Stage
FROM alpine:latest

# Cài đặt ca-certificates để hỗ trợ gọi API HTTPS và tzdata để đồng bộ múi giờ
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

# Copy file thực thi từ stage builder
COPY --from=builder /app/asset-service .

# Port mặc định cho Asset Service
EXPOSE 8002

# Chạy service
CMD ["./asset-service"]