FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY auth/go.mod auth/go.sum ./auth/

RUN cd auth && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY auth/ ./auth/

WORKDIR /workspace/auth

RUN go build -o auth_bin ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /workspace/auth/auth_bin /app/main/

RUN chmod +x /app/main/auth_bin

CMD ["./main/auth_bin"]
