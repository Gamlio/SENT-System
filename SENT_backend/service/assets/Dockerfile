FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/

COPY service/assets/ ./service/assets/


RUN CGO_ENABLED=0 GOOS=linux go build -o asset-service service/assets/cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/asset-service .

EXPOSE 8002

CMD ["./asset-service"]