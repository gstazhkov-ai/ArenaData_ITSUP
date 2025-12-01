# Этап сборки
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Копируем go.mod (и go.sum если есть) и качаем зависимости
COPY go.mod ./
# Если есть go.sum, раскомментируйте: COPY go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN go build -o main .

# Финальный образ (минимальный размер)
FROM alpine:latest
WORKDIR /root/

# Копируем бинарник и шаблоны
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates

# Создаем папку для загрузок
RUN mkdir uploads

# Открываем порт
EXPOSE 8080

# Запуск
CMD ["./main"]