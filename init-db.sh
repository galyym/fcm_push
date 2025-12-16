#!/bin/bash
set -e

# Создаём базу данных если она не существует
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    SELECT 'CREATE DATABASE fcm_push_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'fcm_push_db')\gexec
EOSQL

echo "Database fcm_push_db is ready!"
