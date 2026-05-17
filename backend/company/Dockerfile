FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY company/go.mod company/go.sum ./company/

RUN cd company && go mod download

COPY contracts/ ./contracts/
COPY company/ ./company/

WORKDIR /workspace/company

RUN go build -o company_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /workspace/company/company_bin /app/main/

RUN chmod +x /app/main/company_bin

CMD ["./main/company_bin"]
