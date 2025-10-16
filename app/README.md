# Реализация микросервисной архитектуры

## Структура api

```
api
├── api/                        # Protobuf определения
│   ├── gateway/       
│   │   ├── swagger/    
│   ├── auth/           
│   │   └── auth.proto
│   ├── application/    
│   │   └── application.proto
│   └── log/            
│       └── log.proto
├── cmd/                        # Точки входа для каждого сервиса
│   ├── gateway/        
│   ├── user/           
│   ├── issue/         
│   └── log/            
├── internal/                   # Приватный код, не экспортируемый наружу
│   ├── gateway/                # Реализация бизнес-логики API Gateway
│   │   ├── handlers/   
│   │   ├── services/  
│   │   └── models/     
│   ├── user/                   # Реализация бизнес-логики user_service
│   │   ├── handlers/   
│   │   ├── services/   
│   │   ├── repository/ 
│   │   └── models/     
│   ├── issue/                  # Реализация бизнес-логики issue_service
│   │   ├── handlers/   
│   │   ├── services/   
│   │   ├── repository/ 
│   │   └── models/     
│   └── log/                    # Реализация бизнес-логики log_service
│       ├── handlers/  
│       ├── services/   
│       ├── repository/ 
│       └── models/     
├── pkg/                        # Публичные библиотеки, используемые несколькими сервисами
│   ├── logger/         
│   ├── errors/         
│   └── utils/          
├── protos/                     # Скомпилированные Go-файлы из .proto
│   ├── user/
│   │   └── user.pb.go
│   ├── issue/
│   │   └── issue.pb.go
│   └── log/
│       └── log.pb.go
├── configs/                    # Файлы конфигурации (YAML, JSON, envbbbb)
│   ├── gateway.yaml
│   ├── user.yaml
│   ├── issue.yaml
│   └── log.yaml
├── Makefile                    # Для сборки, запуска, генерации кода
├── go.mod             
├── go.sum
└── README.md
```

## Генерация gRPC из Protobuf 

Для генерации кода необходимо:

1) установить компилятор 
```
https://github.com/protocolbuffers/protobuf/releases
```

2) Распаковать архив в папку и добавить в переменные среды путь к папке bin в папке с компилятором
```
path: folder-with-proto-compiler/bin
```

3) Установить бинарники для компиляции
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4) Сгенерировать прото файлы с помощью Makefile:
```
make proto-gen
```
5) Или консольной командой
```
protoc --proto_path=. --proto_path=api --go_out=protos --go_opt=paths=source_relative --go-grpc_out=protos --go-grpc_opt=paths=source_relative pt=paths=source_relative auth_Service/auth.proto
```

