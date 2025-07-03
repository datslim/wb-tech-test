Микросервис для обработки заказов

### Структура проекта

```
wb-tech-test/
├── cmd/
│   ├── api/          # API сервер
│   └── webserver/    # Веб-сервер для статики
├── internal/
│   ├── api/          # HTTP API
│   ├── cache/        # In-memory кэш
│   ├── db/           # Работа с БД
│   ├── kafka/        # Kafka consumer
│   ├── model/        # Модели данных
│   └── webserver/    # Статический веб-сервер
├── frontend/         # Веб-интерфейс
├── migration/        # SQL миграции
├── docker-compose.yml
└── README.md
```

### Запуск через Makefile

```bash
make run
```

### Миграция базы данных

```bash
make migrate-up
```
### Проверка работы

- **Получение данных через API**: http://localhost:8081/order/<order_uid>
- **Получение данных через веб интерфейс**: http://localhost:3000
- **PgAdmin**: http://localhost:5050

## API Endpoints

### Получение заказа по UID
```bash
GET http://localhost:8081/order/<order_uid>
```

### Отправка заказа в кафку
```
POST http://localhost:8081/orders
```

## Особенности реализации

- **Атомарное сохранение заказов** - все части заказа (order, delivery, payment, items) сохраняются в одной транзакции
- **In-memory кэш** - быстрый доступ к заказам с автоматическим восстановлением из БД при старте
- **Health checks** - проверка готовности зависимостей (PostgreSQL, Kafka)
- **Makefile и Docker** - удобный запуск приложения

## Переменные окружения

Основные переменные в `.env` файле:

- `PG_USER`, `PG_PASS`, `PG_HOST`, `PG_PORT`, `PG_DB` - настройки PostgreSQL
- `API_PORT` - порт API сервера (по умолчанию 8081)
- `STATIC_PORT` - порт веб-интерфейса (по умолчанию 3000)
- `PG_ADMIN_EMAIL`, `PG_ADMIN_PASS`, `PG_ADMIN_PORT` - настройки PgAdmin
