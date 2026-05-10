# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy dependencies từ thư mục gốc backend
COPY go.mod go.sum ./
RUN go mod download

# Copy thư mục pkg dùng chung (Models, DB, Utils)
COPY pkg/ ./pkg/

# Copy mã nguồn riêng của Incident Service
COPY service/incidents/ ./service/incidents/

# Build binary từ thư mục cmd của service
RUN CGO_ENABLED=0 GOOS=linux go build -o incident-service ./service/incidents/cmd/
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/incident-service .

# Tạo thư mục cho bằng chứng số (Evidences)
RUN mkdir -p uploads/incidents uploads/audits

# Port khớp với docker-compose
EXPOSE 8005

CMD ["./incident-service"]