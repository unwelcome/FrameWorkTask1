FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o application_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/application_bin /app/main/

RUN chmod +x /app/main/application_bin

CMD ["./main/application_bin"]