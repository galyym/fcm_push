# FCM Push Notification Service

Микросервис для отправки push-уведомлений через Firebase Cloud Messaging (FCM).

## Возможности

- ✅ Отправка одиночных push-уведомлений
- ✅ Batch отправка (до 500 уведомлений за раз)
- ✅ Поддержка Android и iOS
- ✅ Настройка приоритета уведомлений
- ✅ Дополнительные данные (data payload)
- ✅ API аутентификация
- ✅ Health check endpoint
- ✅ Graceful shutdown
- ✅ Docker support

## Установка и запуск

### Предварительные требования

1. Go 1.23 или выше
2. Firebase проект с настроенным FCM
3. JSON файл с credentials от Firebase

### Шаг 1: Получение Firebase Credentials

1. Перейдите в [Firebase Console](https://console.firebase.google.com/)
2. Выберите ваш проект
3. Перейдите в Project Settings → Service Accounts
4. Нажмите "Generate New Private Key"
5. Сохраните JSON файл

### Шаг 2: Установка зависимостей

```bash
go mod download
```

### Шаг 3: Настройка переменных окружения

Скопируйте `.env.example` в `.env` и заполните:

```bash
cp .env.example .env
```

Отредактируйте `.env`:
```env
SERVER_PORT=8080
FCM_CREDENTIALS_PATH=/path/to/your/firebase-credentials.json
FCM_PROJECT_ID=your-project-id
API_KEY=your-secret-key
```

### Шаг 4: Запуск

```bash
export $(cat .env | xargs)

go run cmd/server/main.go
```

## API Endpoints

### Health Check

```bash
GET /health
```

Ответ:
```json
{
  "status": "ok",
  "service": "fcm-push-service"
}
```

### Отправка одного push-уведомления

```bash
POST /api/v1/push/send
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "token": "device_fcm_token",
  "title": "Новый заказ",
  "body": "У вас новый заказ на поездку",
  "data": {
    "order_id": "12345",
    "type": "new_order"
  },
  "priority": "high",
  "client_id": "driver_123"
}
```

Ответ:
```json
{
  "success": true,
  "message_id": "projects/myproject/messages/0:1234567890"
}
```

### Batch отправка

```bash
POST /api/v1/push/send-batch
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "notifications": [
    {
      "token": "token1",
      "title": "Уведомление 1",
      "body": "Текст 1",
      "priority": "high"
    },
    {
      "token": "token2",
      "title": "Уведомление 2",
      "body": "Текст 2"
    }
  ]
}
```

Ответ:
```json
{
  "success_count": 2,
  "failure_count": 0,
  "results": [
    {
      "success": true,
      "message_id": "msg_id_1"
    },
    {
      "success": true,
      "message_id": "msg_id_2"
    }
  ]
}
```

## Docker

### Сборка образа

```bash
docker build -t fcm-push-service .
```

### Запуск контейнера

```bash
docker run -d \
  -p 8080:8080 \
  -e FCM_CREDENTIALS_PATH=/credentials/firebase.json \
  -e FCM_PROJECT_ID=your-project-id \
  -e API_KEY=your-api-key \
  -v /path/to/firebase.json:/credentials/firebase.json \
  --name fcm-push \
  fcm-push-service
```

## Примеры использования

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/push/send \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "your_device_token",
    "title": "Тест",
    "body": "Тестовое уведомление",
    "priority": "high"
  }'
```