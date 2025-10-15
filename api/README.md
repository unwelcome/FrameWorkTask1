# Реализация микросервисной архитектуры

## Структура api

```
.
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
├── configs/                    # Файлы конфигурации (YAML, JSON, env)
│   ├── gateway.yaml
│   ├── user.yaml
│   ├── issue.yaml
│   └── log.yaml
├── Makefile                    # Для сборки, запуска, генерации кода
├── go.mod             
├── go.sum
└── README.md
```