## Генерация gRPC из Protobuf 

Для генерации кода необходимо:

1) Установить компилятор 
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
protoc --proto_path=api/auth --go_out=auth/api --go_opt=paths=source_relative --go-grpc_out=auth/api --go-grpc_opt=paths=source_relative auth.proto

```