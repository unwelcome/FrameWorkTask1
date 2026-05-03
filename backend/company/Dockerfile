FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o company_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/company_bin /app/main/

RUN chmod +x /app/main/company_bin

CMD ["./main/company_bin"]