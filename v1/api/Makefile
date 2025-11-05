.PHONY: proto-gen

proto-gen:
	@echo "Generating protobuf code..."
	protoc \
	  --proto_path=. \
	  --go_out=generated \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=generated \
	  --go-grpc_opt=paths=source_relative \
	  auth/auth.proto
	@echo "Protobuf code generated"

# Запуск: make proto-gen

# Генерация прото файлов вручную
# protoc --proto_path=. --go_out=generated --go_opt=paths=source_relative --go-grpc_out=generated --go-grpc_opt=paths=source_relative auth/auth.proto

