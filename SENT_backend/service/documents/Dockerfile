# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Tận dụng cache layer cho dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg (Models, DB dùng chung) 
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Document Service
COPY service/documents/ ./service/documents/

# Build binary từ thư mục cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o document-service ./service/documents/cmd/

# Stage 2: Runtime
FROM alpine:latest

# Cài đặt công cụ hỗ trợ xử lý file và mạng
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/document-service .

# Tạo thư mục lưu trữ tài liệu bên trong container
RUN mkdir -p uploads/documents uploads/pdf_previews

# Port khớp với docker-compose
EXPOSE 8007

CMD ["./document-service"]