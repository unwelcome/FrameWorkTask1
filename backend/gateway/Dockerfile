FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY gateway/go.mod gateway/go.sum ./gateway/

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd gateway && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY gateway/ ./gateway/

WORKDIR /workspace/gateway

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o gateway_bin ./cmd/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/gateway/gateway_bin /gateway_bin

CMD ["/gateway_bin"]
