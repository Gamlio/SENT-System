# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/


COPY service/ai/ ./service/ai/

RUN CGO_ENABLED=0 GOOS=linux go build -o ai-service service/ai/cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/ai-service .

# Mang thư mục tài liệu sang môi trường chạy thực tế
COPY --from=builder /app/service/ai/doc/ ./doc/

EXPOSE 8008

CMD ["./ai-service"]