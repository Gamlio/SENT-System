
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
RUN apk add --no-cache pandoc
COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/

COPY service/documents/ ./service/documents/

RUN CGO_ENABLED=0 GOOS=linux go build -o document-service ./service/documents/cmd/

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

COPY --from=builder /app/document-service .

RUN mkdir -p uploads/documents uploads/pdf_previews

EXPOSE 8007

CMD ["./document-service"]