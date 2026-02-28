# ---- Билд стадия ----
FROM golang:1.24-alpine AS builder

# Устанавливаем gcc и musl-dev для CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 👇 ВАЖНО: CGO_ENABLED=1 для SQLite!
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/app/main.go

# ---- Финальная стадия ----
FROM alpine:latest

# Устанавливаем ca-certificates и sqlite библиотеки
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/main .

# Копируем web папку
COPY --from=builder /app/web ./web

# Копируем .env если есть
COPY --from=builder /app/.env ./.env

# Открываем порт
EXPOSE 8080

# Запускаем
CMD ["./main"]
