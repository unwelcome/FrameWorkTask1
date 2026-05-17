module github.com/unwelcome/FrameWorkTask1/backend/auth

go 1.25.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.16.0
	github.com/rs/zerolog v1.35.1
	github.com/unwelcome/FrameWorkTask1/backend/contracts v0.0.0
	github.com/unwelcome/FrameWorkTask1/backend/shared v0.0.0
	golang.org/x/crypto v0.50.0
	golang.org/x/net v0.53.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
)

replace github.com/unwelcome/FrameWorkTask1/backend/contracts => ../contracts

replace github.com/unwelcome/FrameWorkTask1/backend/shared => ../shared
