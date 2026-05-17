FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY company/go.mod company/go.sum ./company/

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd company && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY company/ ./company/

WORKDIR /workspace/company

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o company_bin ./cmd/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/company/company_bin /company_bin

CMD ["/company_bin"]
