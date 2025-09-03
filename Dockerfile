FROM golang:1.22-alpine AS builder

# Install git and CA certificates for secure module download
RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка FFmpeg и CA сертификатов
RUN apk add --no-cache ffmpeg ca-certificates && update-ca-certificates

# Копирование бинарного файла из builder
COPY --from=builder /app/app .

# Копирование конфигурационных файлов
COPY --from=builder /app/configs ./configs

# Создание директорий для данных
RUN mkdir -p ./data/audio

# Установка переменных окружения
ENV APP_ENV=production

# Запуск приложения
CMD ["./app"]