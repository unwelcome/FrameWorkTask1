FROM golang:1.25.1-alpine AS builder

WORKDIR /workspace

COPY contracts/go.mod contracts/go.sum ./contracts/
COPY shared/go.mod shared/go.sum ./shared/
COPY application/go.mod application/go.sum ./application/

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd application && go mod download

COPY contracts/ ./contracts/
COPY shared/ ./shared/
COPY application/ ./application/

WORKDIR /workspace/application

RUN --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o application_bin ./cmd/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/application/application_bin /application_bin

CMD ["/application_bin"]
