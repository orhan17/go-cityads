#!/bin/sh

# Здесь пингуем mysql сервер и пытаемся подключиться
echo "Ожидание MySQL..."
until mysqladmin ping -h "$DB_HOST" --silent; do
  echo "Ждем MySQL..."
  sleep 5
done

echo "MySQL запущен, стартуем API!"
exec "$@"
