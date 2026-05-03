.PHONY: proto-gen proto-gen-auth proto-gen-company proto-gen-application swagger-gen

## proto-gen: Generate protobuf code for all services
proto-gen: proto-gen-auth proto-gen-company proto-gen-application

## proto-gen-auth: Generate protobuf code for auth service
proto-gen-auth:
	@echo "Generating protobuf code for auth service..."
	protoc \
	  --proto_path=auth/api \
	  --go_out=auth/api/generated \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=auth/api/generated \
	  --go-grpc_opt=paths=source_relative \
	  auth.proto
	@echo "Auth service protobuf code generated"

## proto-gen-company: Generate protobuf code for company service
proto-gen-company:
	@echo "Generating protobuf code for company service..."
	protoc \
	  --proto_path=company/api \
	  --go_out=company/api/generated \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=company/api/generated \
	  --go-grpc_opt=paths=source_relative \
	  company.proto
	@echo "Company service protobuf code generated"

## proto-gen-application: Generate protobuf code for application service
proto-gen-application:
	@echo "Generating protobuf code for application service..."
	protoc \
	  --proto_path=application/api \
	  --go_out=application/api/generated \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=application/api/generated \
	  --go-grpc_opt=paths=source_relative \
	  application.proto
	@echo "Application service protobuf code generated"

## swagger-gen: Generate swagger docs for gateway
swagger-gen:
	@echo "Generating swagger docs..."
	cd gateway && swag init -o ./api/docs --dir ./cmd,./internal/entities,./internal/errors,./internal/handlers
	@echo "Swagger docs generated"

# Usage:
#   make proto-gen              - generate all proto contracts
#   make proto-gen-auth         - generate only auth contracts
#   make proto-gen-company      - generate only company contracts
#   make proto-gen-application  - generate only application contracts
#   make swagger-gen            - generate swagger docs for gateway
