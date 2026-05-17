FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY auth/go.mod auth/go.sum ./auth/

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd auth && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY auth/ ./auth/

WORKDIR /workspace/auth

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o auth_bin ./cmd/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/auth/auth_bin /auth_bin

CMD ["/auth_bin"]
