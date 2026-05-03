FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth_bin /app/main/
# COPY --from=builder /app/configs /app/configs

RUN chmod +x /app/main/auth_bin

CMD ["./main/auth_bin"]