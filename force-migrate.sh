#!/bin/bash

# Загружаем переменные из .env
if [ -f .env ]; then
    # Используем grep чтобы пропустить пустые строки и комментарии
    export $(grep -v '^#' .env | xargs)
fi

DB_PASS=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-fcm_push_db}
CONTAINER_NAME="fcm-push-postgres"

echo "🚀 Начинаем принудительную миграцию..."

# 1. Проверяем доступность контейнера
if ! docker ps | grep -q $CONTAINER_NAME; then
    echo "❌ Ошибка: Контейнер $CONTAINER_NAME не запущен!"
    exit 1
fi

# 2. Обновляем пароль пользователя postgres внутри БД на тот, что в .env
# Это решает проблему 'password authentication failed'
echo "🔑 Сихронизируем пароль пользователя postgres в базе..."
docker exec -i $CONTAINER_NAME psql -U postgres -c "ALTER USER postgres WITH PASSWORD '$DB_PASS';"

# 3. Создаем базу если вдруг ее нет (на всякий случай)
docker exec -i $CONTAINER_NAME psql -U postgres -c "SELECT 'CREATE DATABASE $DB_NAME' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec"

# 4. Применяем SQL файл с созданием таблицы
echo "📝 Создаем таблицы и индексы в базе $DB_NAME..."
docker exec -i $CONTAINER_NAME psql -U postgres -d $DB_NAME < migrations/001_create_push_queue.up.sql

echo "✅ Готово! Таблица создана, индексы добавлены, пароль актуализирован."
echo "🔄 Теперь перезапустите сервис: docker-compose restart fcm-push-service"
