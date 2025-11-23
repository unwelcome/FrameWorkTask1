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
	protoc \
	  --proto_path=api/company \
	  --go_out=company/api \
	  --go_out=gateway/api/company \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=company/api \
	  --go-grpc_out=gateway/api/company \
	  --go-grpc_opt=paths=source_relative \
	  company.proto
	@echo "Company service protobuf code generated"
	protoc \
	  --proto_path=api/application \
	  --go_out=application/api \
	  --go_out=gateway/api/application \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=application/api \
	  --go-grpc_out=gateway/api/application \
	  --go-grpc_opt=paths=source_relative \
	  application.proto
	@echo "Application service protobuf code generated"

# Запуск: make proto-gen

# Генерация прото файлов для auth сервиса
# protoc --proto_path=api/auth --go_out=auth/api --go_out=gateway/api/auth --go_opt=paths=source_relative --go-grpc_out=auth/api --go-grpc_out=gateway/api/auth --go-grpc_opt=paths=source_relative auth.proto

# Генерация прото файлов для company сервиса
# protoc --proto_path=api/company --go_out=company/api --go_out=gateway/api/company --go_opt=paths=source_relative --go-grpc_out=company/api --go-grpc_out=gateway/api/company --go-grpc_opt=paths=source_relative company.proto

# Генерация прото файлов для application сервиса
# protoc --proto_path=api/application --go_out=application/api --go_out=gateway/api/application --go_opt=paths=source_relative --go-grpc_out=application/api --go-grpc_out=gateway/api/application --go-grpc_opt=paths=source_relative application.proto
