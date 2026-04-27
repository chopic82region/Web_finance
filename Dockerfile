# --- СТАДИЯ СБОРКИ ---
FROM golang:1.25-alpine AS builder

# Устанавливаем системные зависимости
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем файлы зависимостей для кэширования слоев
COPY go.mod go.sum ./

# Скачиваем зависимости из интернета
RUN go mod download

# Копируем остальные исходники
COPY . .

# Собираем бинарник. Указываем путь к вашему main.go
# Флаг -o задает имя выходного файла
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/main.go

# --- ФИНАЛЬНЫЙ ОБРАЗ ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем скомпилированный сервер и папку с миграциями
COPY --from=builder /app/server .
COPY --from=builder /app/internal/repository/migrations ./migrations

# Порт приложения
EXPOSE 8080

CMD ["./server"]