# User Service

Микросервис для управления пользователями. Часть e-commerce архитектуры на основе gRPC.

## Архитектура

```
cmd/main.go          ← Точка входа
internal/
  ├── app/           ← Инициализация зависимостей
  ├── domain/        ← Бизнес логика (валидация)
  ├── usecase/       ← Оркестрирование (domain + adapter)
  ├── handler/       ← gRPC handlers
  └── adapters/      ← Реалистичные реализации
      ├── storage/   ← Repository (БД)
      └── event/     ← Event Bus (Kafka)
```

## Зависимости

Установите необходимые пакеты:

```bash
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/segmentio/kafka-go
go get github.com/google/uuid
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

## Запуск локально

### 1. Запустите PostgreSQL

```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=ecommerce \
  -p 5432:5432 \
  postgres:15
```

### 2. Примените миграции

```bash
psql -U postgres -d ecommerce -h localhost < ../../migrations/001_create_users_table.sql
```

### 3. Запустите Kafka (опционально)

```bash
docker run -d \
  --name kafka \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -p 9092:9092 \
  confluentinc/cp-kafka:7.5.0
```

### 4. Запустите сервис

```bash
go run ./cmd/main.go
```

Сервис запустится на `localhost:50051`

## Proto файлы

Proto файлы находятся в `/proto` и используют относительные импорты:

```proto
import "common/common.proto";
import "user/v1/user.proto";
```

## Тестирование

### С помощью grpcurl

```bash
# Создать пользователя
grpcurl -plaintext \
  -d '{"email":"test@example.com"}' \
  localhost:50051 user.v1.UserService/CreateUser

# Получить пользователя
grpcurl -plaintext \
  -d '{"id":"<user_id>"}' \
  localhost:50051 user.v1.UserService/GetUser

# Удалить пользователя
grpcurl -plaintext \
  -d '{"id":"<user_id>"}' \
  localhost:50051 user.v1.UserService/DeleteUser
```

## Структура кода объяснена

### Domain (domain/user.go)
Содержит бизнес логику и валидацию. **Не зависит ни от чего** - это "чистый код".

```go
user, err := domain.NewUser(email)  // Валидация встроена сюда
```

### UseCase (usecase/user_service.go)
Оркестрирует domain и adapters. Решает "как это работает".

```go
user, err := domain.NewUser(email)       // Валидация
err := s.repo.Save(ctx, user)            // Сохраняем
err := s.eventBus.PublishUserCreated()   // Публикуем событие
```

### Handler (handler/user_handler.go)
gRPC handler - просто преобразователь запросов в domain объекты.

```go
user, err := h.service.CreateUser(ctx, req.Email)  // Вызываем usecase
return &pb.CreateUserResponse{User: toPb(user)}    // Преобразуем ответ
```

### Adapters
- **storage/postgres** - реализация UserRepository (сохраняет в БД)
- **event/kafka** - реализация UserEventBus (публикует события в Kafka)

## Добавление нового метода

### 1. Обновить proto файл (proto/user/v1/user.proto)
```proto
service UserService {
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}
```

### 2. Сгенерировать код
```bash
protoc -I proto --go_out=pb --go-grpc_out=pb \
  proto/user/v1/user.proto
```

### 3. Добавить в domain (если нужна валидация)
```go
// domain/user.go
func NewListQuery(pageSize int) (*ListQuery, error) {
  if pageSize > 100 {
    return nil, errors.New("page size too large")
  }
  return &ListQuery{pageSize}, nil
}
```

### 4. Добавить в usecase
```go
// usecase/user_service.go
func (s *UserService) ListUsers(ctx context.Context, pageSize int) ([]*domain.User, error) {
  users, err := s.repo.List(ctx, pageSize)
  // ...
}
```

### 5. Добавить в adapter
```go
// adapters/storage/postgres/postgres.go
func (r *PostgresRepository) List(ctx context.Context, pageSize int) ([]*domain.User, error) {
  // SQL код
}
```

### 6. Добавить в handler
```go
// handler/user_handler.go
func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
  users, err := h.service.ListUsers(ctx, int(req.PageSize))
  // ...
}
```

## Environment переменные

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/ecommerce
KAFKA_BROKERS=localhost:9092
GRPC_PORT=50051
LOG_LEVEL=info
```
