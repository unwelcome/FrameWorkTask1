FROM golang:1.24-alpine

# Устанавливаем рабочий каталог
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /app/main ./cmd/api

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]