.PHONY: proto-gen

proto-gen:
	@echo "Generating protobuf code..."
	@mkdir -p protos # Убедимся, что директория protos существует (хотя paths=source_relative часто создает их)
	protoc \
	  --proto_path=api \
	  --proto_path=. \
	  --go_out=. \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=. \
	  --go-grpc_opt=paths=source_relative \
	  api/auth_service/auth.proto
	@echo "Protobuf code generated."

# Запуск: make proto-gen

# protoc --proto_path=. --proto_path=api --go_out=protos --go_opt=paths=source_relative --go-grpc_out=protos --go-grpc_opt=paths=source_relative auth_Service/auth.proto