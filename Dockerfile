# --- builder stage ---
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# --- stage для тестов ---
FROM builder AS test
# При запуске этого контейнера выполняется команда запуска тестов
CMD ["go", "test", "-v", "./..."]

# --- stage для сборки бинарника ---
FROM builder AS final
RUN go build -o geo_offers

# --- финальный образ ---
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    default-mysql-client \
    ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=final /app/geo_offers .
COPY .env .
COPY wait-for-mysql.sh .

RUN chmod +x wait-for-mysql.sh

EXPOSE 3000

CMD ["./wait-for-mysql.sh", "./geo_offers"]
