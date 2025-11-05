.PHONY: proto-gen

proto-gen:
	@echo "Generating protobuf code..."
	protoc \
	  --proto_path=api/auth \
	  --go_out=auth/api \
	  --go_out=gateway/api/auth \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=auth/api \
	  --go-grpc_out=gateway/api/auth \
	  --go-grpc_opt=paths=source_relative \
	  auth.proto
	@echo "Auth service protobuf code generated"

# Запуск: make proto-gen

# Генерация прото файлов для auth сервиса
# protoc --proto_path=api/auth --go_out=auth/api --go_out=gateway/api/auth --go_opt=paths=source_relative --go-grpc_out=auth/api --go-grpc_out=gateway/api/auth --go-grpc_opt=paths=source_relative auth.proto

