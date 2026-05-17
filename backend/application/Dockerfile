FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY application/go.mod application/go.sum ./application/

RUN cd application && go mod download

COPY contracts/ ./contracts/
COPY application/ ./application/

WORKDIR /workspace/application

RUN go build -o application_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /workspace/application/application_bin /app/main/

RUN chmod +x /app/main/application_bin

CMD ["./main/application_bin"]
