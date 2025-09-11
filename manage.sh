#!/bin/bash
set -e

CMD=$1

case "$CMD" in
  start)
    echo "🧹 Чищу старые контейнеры и тома..."
    docker compose down -v

    echo "🚀 Запускаю db..."
    docker compose up -d db

    echo "⚙️ Собираю и запускаю app (с логами)..."
    docker compose build app
    docker compose up app
    ;;

  restart)
    echo "♻️ Пересобираю app и запускаю (с логами)..."
    docker compose build app
    docker compose up app --no-deps
    ;;

  clear)
    echo "🧹 Удаляю контейнеры и тома..."
    docker compose down -v

    echo "🚀 Поднимаю db..."
    docker compose up -d db
    ;;

  *)
    echo "Использование: $0 {start|restart|clear}"
    exit 1
    ;;
esac
