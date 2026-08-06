# Order-Product Service

Микросервис управления продуктами, заказами и подписками в экосистеме Gargantua Labs.

## Обзор

Сервис отвечает за:
- Каталог продуктов (подписки и бандлы)
- Оформление заказов со списанием средств через [balance-service](../balance)
- Управление активными подписками пользователей
- Автоматическое восстановление "зависших" заказов через фоновые воркеры

## Архитектура

```
cmd/order-product/main.go          # Точка входа
internal/
├── config/                        # Конфигурация приложения (YAML + .env)
├── domain/                        # Доменные модели (Product, Order, Subscription)
├── service/                       # Бизнес-логика
├── server/grpc/
│   ├── app/                       # Инициализация gRPC сервера
│   └── handlers/                  # gRPC хендлеры + маппинг proto ↔ domain
├── repository/postgres/
│   ├── order/                     # Репозиторий заказов
│   ├── product/                   # Репозиторий продуктов
│   ├── subscription/              # Репозиторий подписок
│   └── subscription_activation/  # Репозиторий активаций подписок
├── repository/redis_cache/        # Кеширование продуктов в Redis
├── client/                        # gRPC клиент к balance-service
├── infrastructure/postgres/       # Unit of Work (транзакции)
├── middleware/grpc/                # Интерцепторы (error, panic recovery, logging)
├── app_error/                     # Типизированные ошибки приложения
├── telemetry/                     # OpenTelemetry метрики и трейсинг
└── worker/                        # Фоновые воркеры
```

## Стек технологий

| Компонент      | Технология                                         |
|----------------|----------------------------------------------------|
| Язык           | Go 1.25                                            |
| Транспорт      | gRPC (protobuf)                                    |
| База данных    | PostgreSQL (pgx/v5)                                |
| Кеш            | Redis (go-redis/v9)                                |
| Миграции       | goose                                              |
| Трейсинг       | OpenTelemetry (gRPC)                               |
| Логирование    | slog (структурированное)                           |
| Конфигурация   | YAML + .env (cleanenv)                             |

## Схема данных

### Products

| Поле            | Тип          | Описание                                     |
|-----------------|--------------|----------------------------------------------|
| id              | UUID         | Первичный ключ                               |
| type            | product_type | `SUBSCRIPTION` или `BUNDLE`                  |
| name            | VARCHAR(255) | Название                                     |
| description     | TEXT         | Описание                                     |
| duration_days   | INT          | Длительность подписки в днях                 |
| price_amount    | BIGINT       | Цена (в копейках/центах)                     |
| price_currency  | VARCHAR(10)  | Валюта (`RUB`)                               |
| is_active       | BOOLEAN      | Доступен для покупки                         |
| is_archived     | BOOLEAN      | Архивный                                     |

### Product Components (связь Bundle → дочерние продукты)

| Поле               | Тип   | Описание                              |
|--------------------|-------|---------------------------------------|
| parent_product_id  | UUID  | ID бандла                             |
| child_product_id   | UUID  | ID дочернего продукта-подписки        |

### Orders

| Поле                  | Тип         | Описание                                                       |
|-----------------------|-------------|----------------------------------------------------------------|
| id                    | UUID        | Первичный ключ                                                 |
| user_id               | UUID        | ID пользователя                                                |
| product_id            | UUID        | ID купленного продукта                                         |
| product_name          | VARCHAR(255)| Название продукта (денормализовано)                            |
| product_type          | product_type| Тип продукта                                                   |
| duration_days         | INT         | Длительность                                                   |
| amount                | BIGINT      | Сумма списания                                                 |
| currency              | VARCHAR(10) | Валюта                                                         |
| status                | VARCHAR(50) | Статус заказа                                                  |
| balance_operation_id  | UUID        | ID операции в balance-service                                  |
| paid_at               | TIMESTAMP   | Время оплаты                                                   |
| canceled_at           | TIMESTAMP   | Время отмены                                                   |
| expired_at            | TIMESTAMP   | Время истечения                                                |

**Жизненный цикл заказа:**
```
PENDING → PROCESSING → COMPLETED
                    → FAILED
                    → CANCELLED
                    → EXPIRED
```

Уникальный индекс не позволяет создать более одного заказа в статусе `PENDING` для пары `(user_id, product_id)`.

### Subscriptions

| Поле          | Тип          | Описание                                   |
|---------------|--------------|--------------------------------------------|
| id            | UUID         | Первичный ключ                             |
| user_id       | UUID         | ID пользователя                            |
| product_id    | UUID         | ID продукта                                |
| product_name  | VARCHAR(255) | Название продукта                          |
| expires_at    | TIMESTAMP    | Дата истечения подписки                    |
| status        | VARCHAR(16)  | `ACTIVE`, `EXPIRED`, `CANCELED`            |

Уникальный индекс: только одна активная подписка на пару `(user_id, product_id)`.

### Subscription Activations

Аудиторский журнал активаций/продлений подписок.

| Поле             | Тип        | Описание                              |
|------------------|------------|---------------------------------------|
| id               | UUID       | Первичный ключ                        |
| subscription_id  | UUID       | ID подписки                           |
| order_id         | UUID       | ID заказа, в рамках которого активация|
| duration_days    | INT        | Длительность в днях                   |
| started_at       | TIMESTAMP  | Начало действия                       |
| expires_at       | TIMESTAMP  | Конец действия                        |

## gRPC API

Сервис реализует `OrderProductService_API` из [gargantua_common](https://github.com/Krokozabra213/gargantua_common).

| Метод                    | Описание                                      |
|--------------------------|-----------------------------------------------|
| `PurchaseProduct`        | Купить продукт (создать заказ, списать баланс)|
| `CancelSubscription`     | Отменить активную подписку                    |
| `ListProducts`           | Список продуктов с компонентами бандлов       |
| `ListUserOrders`         | Заказы пользователя с пагинацией              |
| `ListUserSubscriptions`  | Подписки пользователя (с фильтром по статусу) |
| `GetOrder`               | Получить заказ по ID                          |
| `GetProduct`             | Получить продукт с компонентами по ID         |

## Покупка продукта (PurchaseProduct)

1. Валидация: продукт существует, активен, не в архиве
2. Создание заказа в статусе `PENDING`
3. Списание средств через [balance-service](../balance) — идемпотентная операция с ключом `order-product:withdraw:{orderID}`
4. При успешном списании — перевод заказа в `COMPLETED` и активация подписки в рамках одной транзакции (Unit of Work)

### Активация подписки
- Если у пользователя ещё нет активной подписки на этот продукт — создаётся новая с `expires_at = now + duration_days`
- Если уже есть активная — `expires_at` продлевается от текущей даты окончания (или от текущего момента, если подписка истекла)
- Создаётся запись в `subscription_activations`

### Бандлы
При покупке бандла рекурсивно активируются все дочерние продукты-подписки.

## Фоновые воркеры

### Pending Orders Worker
Периодически выбирает батч заказов в статусе `PENDING` и пытается процессить их (списать баланс → активировать подписку).

Используется для заказов, которые не были обработаны синхронно (например, из-за сбоя после списания средств).

Конфигурация:
- `workers.pending_orders_interval` — интервал опроса
- `workers.pending_orders_batch_size` — размер батча

### Recovery Worker
Восстанавливает заказы, зависшие в статусе `PROCESSING` дольше заданного таймаута — переводит их обратно в `PENDING` для повторной обработки.

Конфигурация:
- `workers.recovery_order_interval` — интервал опроса
- `workers.processing_order_timeout` — таймаут после которого заказ считается зависшим
- `workers.recovery_order_batch_size` — размер батча

Оба воркера используют случайный jitter для предотвращения "thundering herd".

## Кеширование (Redis)

Продукты с компонентами кешируются в Redis через паттерн cache-aside с защитой от "cache stampede" (singleflight):
- `ListProducts` — список всех продуктов
- `GetProduct` — отдельный продукт по ID

## Конфигурация

```yaml
app:
  name: "order-product"
  version: "1.0.0"

grpc:
  host: "0.0.0.0"
  port: "50052"

postgres:
  dsn: "postgres://user:pass@localhost:5432/order_product?sslmode=disable"
  max_connections: 20

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  prefix: "order-product"

balance:
  addr: "balance:50051"

workers:
  pending_orders_interval: 30s
  pending_orders_batch_size: 100
  recovery_order_interval: 60s
  processing_order_timeout: 5m
  recovery_order_batch_size: 50

slog:
  level: "info"
  format: "json"

telemetry:
  endpoint: "otel-collector:4317"
  service_name: "order-product"
```

## Запуск

### Локально

```bash
# Установка зависимостей
go mod download

# Применение миграций
go run cmd/migrate/main.go

# Запуск сервиса
ENV=dev go run cmd/order-product/main.go
```

### Docker

```bash
docker-compose up -d
```

## Переменные окружения

| Переменная       | Описание                     | По умолчанию |
|------------------|------------------------------|--------------|
| `ENV`            | Имя конфигурационного файла  | `main`       |

Файл конфигурации ищется по пути `configs/{ENV}.yml`.

## Модульные зависимости

- `github.com/Krokozabra213/common` — только типы (`types.PositiveInt`)
- `github.com/Krokozabra213/gargantua_common` — общая инфраструктура (конфигурация, клиенты, телеметрия, proto-контракты, логгер)
