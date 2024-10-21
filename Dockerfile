# Этап сборки
FROM golang:1.20 AS builder

# Устанавливаем рабочую директорию (она будет создана, если не существует)
WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл для Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o emojify

# Этап выполнения
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Устанавливаем зависимости для работы Go-бинарника на Alpine
RUN apk --no-cache add ca-certificates

# Копируем скомпилированный бинарник
COPY --from=builder /app/emojify .

# Указываем порт, который будет слушать приложение
EXPOSE 8080

# Запуск приложения
CMD ["./emojify"]
