# Memology Backend

API для платформы генерации мемов на Go с JWT авторизацией, PostgreSQL, асинхронной обработкой, MinIO и Clean Architecture.

## Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/lDizil/memology-backend.git
cd memology-backend
```

### 2. Настройка переменных окружения

Скопируйте пример и настройте .env:

```bash
cp .env.example .env
# Откройте .env и укажите свои параметры (DB, JWT, MinIO и др.)
```

### 3. Сборка и запуск контейнеров

```bash
docker-compose up --build
# или
make dev-up
```

Все сервисы (Go backend, PostgreSQL, MinIO) будут запущены автоматически.

Для локальной разработки без Docker:

```bash
go mod tidy
make run
# или
go run ./cmd/server
```

## API Endpoints

### Аутентификация

- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход (username или email)
- `POST /api/v1/auth/refresh` - Обновление токена
- `POST /api/v1/auth/logout` - Выход
- `POST /api/v1/auth/logout-all` - Выход со всех устройств

### Пользователи

#### требует авторизации

- `GET /api/v1/users/profile` - Получить профиль
- `PUT /api/v1/users/profile/update` - Обновить профиль
- `POST /api/v1/users/change-password` - Сменить пароль
- `DELETE /api/v1/users/account` - Удалить аккаунт пользователя

#### не требует авторизации

- `GET /api/v1/users/list` - Список пользователей
- `GET /api/v1/users/profile/:id` - Получить профиль по ID

### Мемы

#### не требует авторизации

- `GET /api/v1/memes` - Получить все мемы
- `GET /api/v1/memes/public` - Публичные мемы с пагинацией и поиском (`?page=1&limit=20&search=текст`)
- `GET /api/v1/memes/styles` - Список доступных стилей генерации
- `GET /api/v1/memes/:id` - Получить мем по ID
- `GET /api/v1/memes/:id/status` - Проверить статус генерации мема

#### требует авторизации

- `POST /api/v1/memes/generate` - Сгенерировать мем (асинхронно)
- `GET /api/v1/memes/my` - Свои мемы с пагинацией и поиском (`?page=1&limit=20&search=текст`)
- `DELETE /api/v1/memes/:id` - Удалить свой мем

## Документация

### Для разработчиков (интерактивная)

Swagger UI доступен по адресу: `http://localhost:8080/api/swagger/index.html`

### Для фронтенда (OpenAPI спецификация)

Был реализован конвертер из Swagger 2.0 в OpenAPI 3.0.3

OpenAPI JSON спецификация: `http://localhost:8080/api/openapi.json`

Используйте этот endpoint для генерации клиентского SDK или типов TypeScript:

```bash
# Пример для TypeScript
npx openapi-typescript http://localhost:8080/openapi.json --output ./types/api.ts
```

Документация автоматически генерируется из аннотаций в коде и содержит все доступные endpoints с примерами запросов.

### Формат ответа с пагинацией

Для роутов с пагинацией возвращается структура:

```json
{
  "memes": [...],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

Параметры запроса:

- `page` — номер страницы (по умолчанию 1)
- `limit` — количество элементов на странице (по умолчанию 20, максимум 100)
- `search` — поиск по промпту (опционально)

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


MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin123
MINIO_ENDPOINT=minio:9000
MINIO_PUBLIC_URL=http://localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_BUCKET=memes

TASK_PROCESSOR_WORKERS=10
TASK_PROCESSOR_QUEUE_SIZE=100
TASK_PROCESSOR_POLL_INTERVAL=5s

# AI сервис (нейронная сеть для генерации мемов)
AI_BASE_URL=http://localhost:7080
AI_TIMEOUT=120s
```

## Особенности

- **Авторизация**: Можно входить как по username, так и по email
- **JWT**: Access (1 час) + Refresh (7 дней) токены в HTTP-only cookies
- **Пароли**: Хешируются через Argon2
- **UUID**: Используются для всех ID
- **GORM**: Auto-миграции БД при старте
- **MinIO**: S3-совместимое хранилище для изображений мемов
- **Clean Architecture**: Разделение на слои handlers → services → repository

- **Асинхронная генерация мемов**: После запроса `/memes/generate` возвращается объект со статусом `pending`. Task Processor автоматически обрабатывает задачу: опрашивает AI-сервис каждые 5 секунд (до 120 попыток), загружает результат в MinIO и обновляет статус на `completed`
- **MinIO**: S3-совместимое хранилище для изображений мемов. Консоль: `http://localhost:9001` (логин: minioadmin, пароль: minioadmin)
- **Task Processor**: Пул воркеров для асинхронной обработки задач генерации. Настраивается через `.env`:
  - `TASK_PROCESSOR_WORKERS` — количество воркеров (по умолчанию 10)
  - `TASK_PROCESSOR_QUEUE_SIZE` — размер очереди задач (по умолчанию 100)
  - `TASK_PROCESSOR_POLL_INTERVAL` — интервал опроса AI-сервиса (по умолчанию 5s)
- **Stuck Tasks Scanner**: Автоматическое восстановление застрявших задач каждый час

## Генерация мемов

### Запрос на генерацию

```bash
POST /api/v1/memes/generate
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "prompt": "я купил компьютер за 1000000",
  "style": "anime",
  "is_public": true
}
```

- `prompt` — текст для генерации (обязательный)
- `style` — стиль генерации (опционально, список стилей: `GET /api/v1/memes/styles`)
- `is_public` — публичный мем или приватный (по умолчанию `true`)

Мем создаётся со статусом `pending`. После начала обработки нейросетью статус меняется на `processing`.

### Асинхронная генерация

1. Мем создаётся со статусом `pending`, возвращается объект с `id` и `task_id`
2. Task Processor автоматически берёт задачу и опрашивает AI-сервис
3. При успехе изображение загружается в MinIO
4. Статус обновляется на `completed`, появляется `image_url`
5. При ошибке статус становится `failed`

**Проверка статуса:** `GET /api/v1/memes/{id}/status` или `GET /api/v1/memes/{id}`

### MinIO Console

Для просмотра загруженных файлов:

- URL: `http://localhost:9001`
- Login: `minioadmin`
- Password: `minioadmin`

## Команды разработки

```bash
make help        # Показать все команды
make build       # Собрать приложение
make run         # Запустить локально
make swagger     # Обновить документацию
make clean       # Очистить всё
# и другие
```

## Деплой на сервер

Настроен автоматический деплой при push в `main` ветку через GitHub Actions.

Подробная инструкция по настройке: [.github/DEPLOY.md](.github/DEPLOY.md)

**Кратко:**

1. Добавьте GitHub Secrets: `SERVER_HOST`, `SERVER_USER`, `SSH_PRIVATE_KEY`
2. Подготовьте сервер (Docker, Git)
3. Push в `main` → автоматический деплой ✅

## Архитектура

- Проект построен по принципам Clean Architecture:

  - handlers (Gin HTTP)
  - services (бизнес-логика)
  - repository (работа с БД)
  - models (GORM-сущности)
  - middleware (JWT, CORS)
  - router (регистрация роутов)
  - config (конфигурация)
  - database (инициализация и миграции)

- Вся логика разделена по слоям для удобства поддержки и масштабирования.
