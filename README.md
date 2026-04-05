# auth-service

`auth-service` — небольшой микросервис аутентификации на Go.

Сервис реализует базовый auth flow для backend-систем и внутренних сервисов:

- регистрация пользователя
- подтверждение email через OTP-код
- логин по email и паролю
- выдача JWT access token
- обновление access token через refresh flow
- logout с отзывом сессии
- получение данных текущего пользователя по эндпоинту `/me`

## Технологии

- Go
- Chi
- PostgreSQL
- Redis
- `pgx`
- `sqlc`
- `goose`
- `golang-jwt/jwt/v5`

## HTTP API

Health check:

```text
GET /health
```

Базовый префикс:

```text
/api/v1/auth
````

### Доступные endpoints

| Метод | Путь        | Назначение                                   |
|-------|-------------|----------------------------------------------|
| POST  | `/register` | Регистрация пользователя                     |
| POST  | `/verify`   | Подтверждение email по OTP-коду              |
| POST  | `/login`    | Логин и получение access token               |
| POST  | `/refresh`  | Обновление access token через refresh cookie |
| POST  | `/logout`   | Отзыв текущей refresh-сессии                 |
| GET   | `/me`       | Данные текущего пользователя                 |

---

## Как работает аутентификация

### Access token

Access token — короткоживущий JWT, который возвращается в теле ответа.

Его нужно передавать в заголовке:

```http
Authorization: Bearer <access_token>
```

### Refresh token

Refresh token хранится в `HttpOnly` cookie.

Он используется в запросах:

* `POST /api/v1/auth/refresh`
* `POST /api/v1/auth/logout`

При каждом успешном refresh refresh token ротируется.

## Структура проекта

```text
cmd/auth-service/main.go        - точка входа
internal/app                    - сборка приложения и wiring зависимостей
internal/auth                   - бизнес-логика аутентификации
internal/config                 - конфигурация из env
internal/logger                 - настройка логгера
internal/otp                    - абстракции для OTP
internal/storage/postgres       - PostgreSQL и SQL-запросы
internal/storage/redis          - Redis и хранение OTP-кодов
internal/transport/http         - HTTP handlers, middleware, router
migrations                      - SQL-миграции
```

---

## Модель хранения

### PostgreSQL

#### Таблица `users`

Содержит auth-данные пользователя:

* `id`
* `email`
* `password_hash`
* `email_verified`
* `created_at`
* `updated_at`
* `deleted_at`

#### Таблица `sessions`

Содержит refresh-сессии:

* `id`
* `user_id`
* `refresh_hash`
* `expires_at`
* `revoked_at`
* `created_at`
* `user_agent`
* `ip`

### Redis

Redis используется для хранения OTP-кодов подтверждения email с TTL.

Пример ключа:

```text
verify_email:<email>
```

## Локальный запуск

### 1. Подготовить `.env`

Пример в `.env.example`

### 2. Поднять Docker Compose

```bash
docker compose up --build
```

### 3. Применить миграции

```bash
make migrate-up
```

## Текущий scope проекта

Реализовано:

* register
* verify email
* login
* refresh
* logout
* me
* PostgreSQL
* Redis
* Docker Compose

Не реализовано:

* полноценная role/permission model
* gRPC transport
* метрики и tracing

## Дальнейшее развитие

* resend verification code
* request ID и request logging middleware
* unit и integration тесты для auth flow и HTTP handlers
* хранение ролей в БД
* метрики и tracing
* интеграция с email sender
* gRPC transport
