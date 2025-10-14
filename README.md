# Memology Backend

API для платформы генерации мемов на Go с JWT авторизацией и PostgreSQL.

## Быстрый старт

### 1. Запуск БД для разработки

```bash
make dev-db
# или
docker-compose -f docker-compose.dev.yml up -d
```

### 2. Запуск приложения

```bash
# Скачать зависимости
go mod tidy

# Запустить приложение
make run
# или
go run ./cmd/server
```

### 3. Полный запуск в Docker

```bash
make dev-up
# или
docker-compose up --build
```

## API Endpoints

### Аутентификация

- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход (username или email)
- `POST /api/v1/auth/refresh` - Обновление токена
- `POST /api/v1/auth/logout` - Выход
- `POST /api/v1/auth/logout-all` - Выход со всех устройств

### Пользователи (требует авторизации)

- `GET /api/v1/users/profile` - Получить профиль
- `PUT /api/v1/users/profile/update` - Обновить профиль
- `POST /api/v1/users/change-password` - Сменить пароль
- `GET /api/v1/users/list` - Список пользователей

## Документация

### Для разработчиков (интерактивная)

Swagger UI доступен по адресу: `http://localhost:8080/swagger/index.html`

### Для фронтенда (OpenAPI спецификация)

OpenAPI JSON спецификация: `http://localhost:8080/swagger/doc.json`

Используйте этот endpoint для генерации клиентского SDK или типов TypeScript:

```bash
# Пример для TypeScript
npx openapi-typescript http://localhost:8080/swagger/doc.json --output ./types/api.ts
```

Документация автоматически генерируется из аннотаций в коде и содержит все доступные endpoints с примерами запросов.

## Переменные окружения

Скопируйте `.env.example` в `.env` и настройте под свои нужды:

```bash
cp .env.example .env
```

Пример переменных (см. `.env.example`):

```bash
SERVER_PORT=8080
SERVER_HOST=localhost

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=memology
DB_SSLMODE=disable

JWT_SECRET=your-very-secret-jwt-key-change-this-in-production
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=168h
```

## Особенности

- **Авторизация**: Можно входить как по username, так и по email
- **JWT**: Access (1 час) + Refresh (7 дней) токены
- **Пароли**: Хешируются через Argon2
- **UUID**: Используются для всех ID
- **GORM**: Auto-миграции БД при старте

## Команды разработки

```bash
make help        # Показать все команды
make build       # Собрать приложение
make run         # Запустить локально
make dev-db      # Запустить только PostgreSQL
make swagger     # Обновить документацию
make clean       # Очистить всё
```
