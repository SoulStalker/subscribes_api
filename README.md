# Subscribes API


REST API для агрегации и управления данными об онлайн-подписках пользователей.

## Быстрый старт

```bash
# Клонировать репозиторий
git clone https://github.com/yourname/subscriptions-service.git
cd subscriptions-service

# Настроить переменные окружения
cp .env.example .env

# Запустить сервис
docker-compose up --build

# API доступен на http://localhost:8080
# Swagger UI: http://localhost:8080/swagger/index.html
```

## Возможности

- ✅ **CRUDL операции** над подписками (Create, Read, Update, Delete, List)
- ✅ **Фильтрация** по user_id и названию сервиса
- ✅ **Расчет стоимости** подписок за выбранный период
- ✅ **PostgreSQL** с автоматическими миграциями
- ✅ **Структурированное логирование** (Zap)
- ✅ **Swagger документация** (автогенерация)
- ✅ **Docker Compose** для локальной разработки
- ✅ **Graceful shutdown** с таймаутом

## Архитектура

```
subscriptions-service/
├── cmd/api/              # Точка входа приложения
├── internal/
│   ├── config/          # Конфигурация (Cleanenv)
│   ├── domain/          # Бизнес-модели
│   ├── handler/.        # HTTP-хендлеры (Gin)
│   ├── service/         # Бизнес-логика
│   ├── repository/      # Работа с БД (pgx)
│   └── logger/          # Настройка логгера
├── migrations/          # SQL-миграции
├── docs/                # Swagger-документация
├── docker-compose.yml
├── Dockerfile
├── config/config.yaml  # Конфигурация приложения c 
```

**Стек технологий**:
- Go 1.25
- Gin (HTTP router)
- PostgreSQL 16
- pgx v5 (PostgreSQL driver)
- Cleanenv (конфигурация)
- Zap (логирование)
- Swaggo (Swagger)
- golang-migrate (миграции)
- Docker & Docker Compose

## Установка и настройка

### Требования

- Docker 24.0+
- Docker Compose 2.20+
- Go 1.25+ (для локальной разработки)

### Локальный запуск без Docker

```bash
# Установить зависимости
go mod download

# Запустить PostgreSQL
docker run -d \
  --name subscriptions-postgres \
  -e POSTGRES_USER=subscriptions_user \
  -e POSTGRES_PASSWORD=strong_password_here \
  -e POSTGRES_DB=subscriptions_db \
  -p 5432:5432 \
  postgres:16-alpine

# Применить миграции
migrate -path migrations \
  -database "postgres://subscriptions_user:strong_password_here@localhost:5432/subscriptions_db?sslmode=disable" \
  up

# Сгенерировать Swagger-документацию
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/api/main.go --output docs

# Запустить сервис
go run cmd/api/main.go
```

## API Endpoints

### Подписки

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/subscriptions` | Создать подписку |
| GET | `/api/v1/subscriptions/:id` | Получить подписку по ID |
| GET | `/api/v1/subscriptions` | Список подписок (с фильтрами) |
| PUT | `/api/v1/subscriptions/:id` | Обновить подписку |
| DELETE | `/api/v1/subscriptions/:id` | Удалить подписку |
| GET | `/api/v1/subscriptions/total-cost` | Рассчитать стоимость |

### Swagger UI

Интерактивная документация доступна по адресу:
```
http://localhost:8080/swagger/index.html
```

## Примеры использования

### Создание подписки

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

**Ответ**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "2025-07-01",
  "end_date": "2025-12-01",
  "created_at": "2025-12-09T10:30:00Z",
  "updated_at": "2025-12-09T10:30:00Z"
}
```

### Получение списка подписок

```bash
# Все подписки пользователя
curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

# Фильтр по названию сервиса
curl "http://localhost:8080/api/v1/subscriptions?service_name=Yandex"

# Комбинированный фильтр
curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Plus"
```

### Получение подписки по ID

```bash
curl http://localhost:8080/api/v1/subscriptions/123e4567-e89b-12d3-a456-426614174000
```

### Обновление подписки

```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus Family",
    "price": 600,
    "start_date": "07-2025",
    "end_date": "12-2026"
  }'
```

### Удаление подписки

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/123e4567-e89b-12d3-a456-426614174000
```

### Расчет суммарной стоимости

```bash
# За весь 2025 год для пользователя
curl "http://localhost:8080/api/v1/subscriptions/total-cost?start_period=2025-01-01&end_period=2025-12-31&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

# Только подписки Yandex
curl "http://localhost:8080/api/v1/subscriptions/total-cost?start_period=2025-01-01&end_period=2025-12-31&service_name=Yandex"

# Без фильтров (все подписки за период)
curl "http://localhost:8080/api/v1/subscriptions/total-cost?start_period=2025-01-01&end_period=2025-12-31"
```

**Ответ**:
```json
{
  "total": 4800
}
```

## Конфигурация

### Переменные окружения (.env)

```bash
# PostgreSQL
DB_HOST=postgres          
DB_PORT=5432             
DB_USER=subscriptions_user
DB_PASSWORD=strong_password_here
DB_NAME=subscriptions_db
```

### Файл конфигурации (config.yaml)

```yaml
server:
  port: 8080            
  mode: debug             # gin mode: debug/release

database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
  user: ${DB_USER:postgres}
  password: ${DB_PASSWORD:secret}
  dbname: ${DB_NAME:subscriptions}
  sslmode: disable
  max_conns: 25          
  max_idle_conns: 5     

log:
  level: info             # debug/info/warn/error
  encoding: json          # json/console
```

## 🗄 База данных

### Схема таблицы subscriptions

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);
```

### Миграции

Миграции применяются автоматически при запуске через Docker Compose.

Для ручного управления:

```bash
# Установить migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Применить миграции
migrate -path migrations \
  -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" \
  up

# Откатить последнюю миграцию
migrate -path migrations \
  -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" \
  down 1

# Создать новую миграцию
migrate create -ext sql -dir migrations -seq add_new_field
```

## Тестирование

### Юнит-тесты

```bash
go test ./... -v -cover
```

### Интеграционные тесты (TODO)

```bash
# С использованием testcontainers
go test ./tests/integration/... -v
```

### Ручное тестирование через Postman

Импортируй Swagger JSON из `http://localhost:8080/swagger/doc.json` в Postman.

## Логирование

Сервис использует структурированное логирование через Zap.

**Примеры логов (JSON)**:
```json
{
  "level": "info",
  "ts": "2025-12-09T10:30:00.000Z",
  "msg": "request",
  "method": "POST",
  "path": "/api/v1/subscriptions",
  "status": 201,
  "latency": "15.2ms"
}
```

**Уровни логирования**:
- `debug` — детальная информация для отладки
- `info` — информационные сообщения (HTTP-запросы, старт/стоп)
- `warn` — предупреждения
- `error` — ошибки выполнения

Настройка уровня в `config.yaml` → `log.level`.

## Docker

### Сборка образа

```bash
docker build -t subscriptions-service:latest .
```

### Запуск контейнера

```bash
docker run -d \
  --name subscriptions-api \
  -p 8080:8080 \
  --env-file .env \
  subscriptions-service:latest
```

### Docker Compose команды

```bash
# Запуск в фоне
docker-compose up -d

# Просмотр логов
docker-compose logs -f api

# Остановка
docker-compose down

# Пересборка и запуск
docker-compose up --build

# Очистка (включая volumes)
docker-compose down -v
```

## Разработка

### Структура проекта (Clean Architecture)

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

**Принципы**:
- Dependency Injection через конструкторы
- Интерфейсы для абстракции слоев
- Ошибки оборачиваются с контекстом (`fmt.Errorf`)
- Контекст передается через все слои

### Добавление нового endpoint

1. Добавь метод в `internal/repository/postgres/subscription.go`
2. Добавь метод интерфейса и реализацию в `internal/service/subscription.go`
3. Добавь хендлер в `internal/handler/http/handler.go`
4. Добавь Swagger-аннотации
5. Регенерируй документацию: `swag init -g cmd/api/main.go`

### Обновление Swagger

```bash
# После изменения аннотаций
swag init -g cmd/api/main.go --output docs

# Форматирование аннотаций
swag fmt
```

## Production-Ready улучшения

### TODO

- [ ] **Тесты**: интеграционные (testcontainers)
- [ ] **Метрики**: Prometheus + Grafana дашборды
- [ ] **Трейсинг**: OpenTelemetry + Jaeger
- [ ] **Rate Limiting**: middleware для защиты от DDoS
- [ ] **Authentication**: JWT-токены для защиты API
- [ ] **CI/CD**: GitHub Actions для автотестов и деплоя
- [ ] **Kubernetes**: Helm-чарты для деплоя
- [ ] **Terraform**: IaC для инфраструктуры (RDS, EKS)
- [ ] **Пагинация**: limit/offset для списков
- [ ] **Валидация**: расширенная валидация входных данных
- [ ] **Кеширование**: Redis для частых запросов

### Мониторинг

Добавить эндпоинт health check:

```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})
```

## Contributing

1. Fork репозиторий
2. Создай feature-ветку (`git checkout -b feature/amazing-feature`)
3. Закоммить изменения (`git commit -m 'Add amazing feature'`)
4. Push в ветку (`git push origin feature/amazing-feature`)
5. Открой Pull Request

## Лицензия

MIT License. См. `LICENSE` для деталей.

## Контакты

- GitHub: [@soulstalker](https://github.com/soulstalker)
- Email: almazwork@gmail.com

---

**Версия**: 1.0.0  
**Go версия**: 1.25+  
**Последнее обновление**: Декабрь 2025
