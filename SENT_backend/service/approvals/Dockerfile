# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependencies từ thư mục gốc backend
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg dùng chung
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Approvals Service
COPY service/approvals/ ./service/approvals/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o approval-service service/approvals/cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/approval-service .

# Port khớp với hệ thống
EXPOSE 8013

CMD ["./approval-service"]