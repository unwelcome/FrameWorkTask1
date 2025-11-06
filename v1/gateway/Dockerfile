FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o gateway_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway_bin /app/main/

RUN chmod +x /app/main/gateway_bin

CMD ["./main/gateway_bin"]