FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY gateway/go.mod gateway/go.sum ./gateway/

RUN cd gateway && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY gateway/ ./gateway/

WORKDIR /workspace/gateway

RUN go build -o gateway_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /workspace/gateway/gateway_bin /app/main/

RUN chmod +x /app/main/gateway_bin

CMD ["./main/gateway_bin"]
