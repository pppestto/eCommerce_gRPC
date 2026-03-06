# E-Commerce Backend (gRPC Microservices)

Бэкенд e-commerce приложения на Go: микросервисная архитектура с gRPC между сервисами, BFF с REST API для клиентов, событийная шина (Kafka), кэш (Redis), метрики и распределённый трейсинг.

---

## О проекте

Многосервисное приложение для типичного сценария интернет-магазина:

- **Регистрация и авторизация** пользователей (JWT).
- **Каталог товаров**: просмотр списка и карточки товара, создание товаров (через gRPC).
- **Оформление заказов**: корзина из нескольких товаров, проверка пользователя и наличия товаров, сохранение заказа и публикация событий.

Сервисы общаются по **gRPC**, клиент (веб/мобильный) — через **REST API** единой точки входа (BFF). События (создание пользователя, товара, заказа) публикуются в **Kafka** для возможной последующей обработки или аналитики.

---

## Реализовано

| Область | Реализация |
|--------|------------|
| **Пользователи** | Регистрация, логин, JWT, хэширование паролей (bcrypt), хранение в Postgres, события в Kafka |
| **Товары** | CRUD по gRPC, хранение в Postgres, публикация событий в Kafka |
| **Заказы** | Создание заказа (проверка user/product), хранение в Postgres, outbox-паттерн для событий в Kafka, consumer для дублирования событий в БД |
| **BFF** | REST: регистрация, логин, создание заказа (с JWT), просмотр списка/карточки товара (без авторизации). Обращения к user/product/order по gRPC, кэш товаров в Redis, circuit breaker на клиентах |
| **Наблюдаемость** | Prometheus-метрики (BFF), структурированное логирование (slog), распределённый трейсинг (OpenTelemetry → Jaeger) по всем сервисам |
| **Инфраструктура** | Docker Compose: Postgres, Redis, Kafka, инициализация топиков, Jaeger, Prometheus; миграции БД при старте |

---

## Стек технологий

- **Язык:** Go
- **RPC / API:** gRPC, Protocol Buffers, REST (net/http)
- **БД:** PostgreSQL (pgx)
- **Кэш:** Redis (go-redis)
- **Очереди/события:** Kafka (segmentio/kafka-go), топики user-events, product-events, order-events
- **Авторизация:** JWT (golang-jwt)
- **Наблюдаемость:** Prometheus (client_golang), OpenTelemetry (SDK + OTLP), Jaeger, slog
- **Устойчивость:** gobreaker (circuit breaker) на gRPC-клиентах в BFF

---

## Архитектура

```
       ┌─────────────────────────────────────────────────────────┐
       │                    BFF (HTTP :8080)                     │   
       │  REST: /api/auth/*, /api/orders, /api/products, /metrics│
       │  JWT middleware, Redis cache, OTEL + Prometheus         │
       └────────────────────────────┬────────────────────────────┘
                                    │ gRPC
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
         ▼                          ▼                          ▼
 ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
 │  user-service   │       │ product-service │       │  order-service  │
 │  :50051         │       │ :50052          │       │  :50053         │
 │  Postgres       │       │ Postgres        │       │  Postgres       │
 │  Kafka (events) │       │ Kafka (events)  │       │  Kafka outbox   │
 │  OTEL → Jaeger  │       │ OTEL → Jaeger   │       │  OTEL → Jaeger  │
 └─────────────────┘       └─────────────────┘       └─────────────────┘
         │                          │                          │
         └──────────────────────────┼──────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │  Kafka  │  Redis  │  Postgres │
                    └───────────────────────────────┘
```

- **BFF** — единственная точка входа для клиента: авторизация, агрегация вызовов к сервисам, кэширование каталога в Redis.
- **user-service** — пользователи, логин/регистрация, события в Kafka.
- **product-service** — каталог товаров, gRPC + события.
- **order-service** — заказы, outbox для надёжной доставки событий в Kafka, consumer для записи событий в БД.

Сервисы следуют слоистой структуре: handler (gRPC) → usecase → domain, адаптеры (postgres, kafka) инжектятся в usecase.

---

## Структура репозитория

```
├── cmd/                          # (точки входа размазаны по services/*/cmd)
├── proto/                        # Protobuf-описания API
│   ├── common/
│   ├── user/v1/
│   ├── product/v1/
│   └── order/v1/
├── pb/                           # Сгенерированный Go-код из proto
├── pkg/otel/                     # Инициализация OpenTelemetry (TracerProvider, OTLP)
├── services/
│   ├── bff/                      # HTTP API, gRPC-клиенты, Redis, middleware
│   ├── user-service/
│   ├── product-service/
│   └── order-service/
├── services/common/              # Общий код: logger, metrics, config
├── migrations/                   # SQL-миграции (Postgres)
├── deployments/                  # Конфиги (Prometheus и т.д.)
├── scripts/                      # E2E-тест (PowerShell), генерация proto
├── docker-compose.yml
├── go.mod / go.sum
├── README.md
└── TESTING.md                    # Инструкции по запуску и проверке
```

Каждый сервис: `cmd/main.go`, `internal/app`, `internal/handler`, `internal/usecase`, `internal/domain`, `internal/adapters` (storage, event/kafka, auth при необходимости).

---

## Быстрый старт

### Требования

- Docker и Docker Compose
- (опционально) Go 1.21+, grpcurl — для локального запуска

### Запуск всего стека

```bash
docker-compose up -d --build
```

После старта (подождать ~30–60 сек, пока поднимутся БД, Kafka, kafka-init, сервисы):

| Сервис | Адрес | Назначение |
|--------|--------|------------|
| BFF (REST) | http://localhost:8080 | Единая точка входа для клиента |
| user-service (gRPC) | localhost:50051 | Пользователи |
| product-service (gRPC) | localhost:50052 | Товары |
| order-service (gRPC) | localhost:50053 | Заказы |
| Jaeger UI | http://localhost:16686 | Трейсы запросов |
| Prometheus | http://localhost:9090 | Метрики |

---

## API (BFF)

- `GET /api/health` — проверка доступности
- `POST /api/auth/register` — регистрация (email, password)
- `POST /api/auth/login` — логин (email, password) → JWT
- `GET /api/products` — список товаров (query: `page`, `size`, `category`)
- `GET /api/products/{id}` — карточка товара
- `POST /api/orders` — создание заказа (требуется JWT), тело: `{"items":[{"product_id","quantity","price":{...}}]}`
- `GET /metrics` — метрики Prometheus

gRPC-сервисы (user, product, order) можно вызывать напрямую через grpcurl при необходимости (см. TESTING.md).

---

## Наблюдаемость

- **Метрики:** BFF отдаёт Prometheus-метрики на `/metrics`.
- **Логи:** структурированные логи (slog) в сервисах.
- **Трейсинг:** OpenTelemetry (OTLP) настроен во всех сервисах; при заданном `OTEL_EXPORTER_OTLP_ENDPOINT` (в docker-compose — Jaeger) трейсы уходят в Jaeger. Один HTTP-запрос в BFF виден как один trace с вложенными gRPC-вызовами.

---


