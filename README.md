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

- `GET /api/v1/memes` - Получить все мемы
- `GET /api/v1/memes/:id` - Получить мем по ID
- `POST /api/v1/memes/generate` - Сгенерировать мем
- `GET /api/v1/memes/my` - Получить свои мемы
- `DELETE /api/v1/memes/:id` - Удалить свой мем
- `GET /api/v1/memes/public` - Получить публичные мемы с пагинацией
- `GET /api/v1/memes/styles` - Получить список доступных стилей генерации
- `GET /api/v1/memes/:id/status` - Проверить статус генерации мемов (асинхронная обработка)
- `GET /api/v1/memes/search/public` - Поиск по публичным мемам
- `GET /api/v1/memes/search/private` - Поиск по личным мемам (требует авторизации)

## Документация

### Для разработчиков (интерактивная)

Swagger UI доступен по адресу: `http://localhost:8080/swagger/index.html`

### Для фронтенда (OpenAPI спецификация)

Был реализован свой конвертатор из 2 спецификации в версию 3.0.3

OpenAPI JSON спецификация: `http://localhost:8080/openapi.json`

Используйте этот endpoint для генерации клиентского SDK или типов TypeScript:

```bash
# Пример для TypeScript
npx openapi-typescript http://localhost:8080/openapi.json --output ./types/api.ts
```

Документация автоматически генерируется из аннотаций в коде и содержит все доступные endpoints с примерами запросов.

### Формат ответа с пагинацией

Для всех роутов с пагинацией (мемы, пользователи) возвращается структура:

```json
{
	"items": [...],
	"total": 123,
	"page": 1,
	"limit": 20
}
```

Пример для мемов:

```json
{
	"items": [ { ...meme... }, ... ],
	"total": 42,
	"page": 1,
	"limit": 20
}
```

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

AI_BASE_URL=http://localhost:7080 (необходимо указать сервер, где запущена нейронная сеть, создающая мемы)
```

## Особенности

- **Авторизация**: Можно входить как по username, так и по email
- **JWT**: Access (1 час) + Refresh (7 дней) токены в HTTP-only cookies
- **Пароли**: Хешируются через Argon2
- **UUID**: Используются для всех ID
- **GORM**: Auto-миграции БД при старте
- **MinIO**: S3-совместимое хранилище для изображений мемов
- **Clean Architecture**: Разделение на слои handlers → services → repository

- **Асинхронная генерация мемов**: После запроса `/memes/generate` возвращается объект с `status: pending` и `task_id`. Для проверки статуса можно использовать `/memes/{id}/status`. По умолчанию запускаются workers, которые начинают проверять статус каждые 5 секунд 120 раз (настраиваться с помощью .env). После завершения статус становится `completed`, появляется `image_url` (ниже -> Task Processor)
- **MinIO**: Хранение изображений мемов и аватаров, публичные ссылки для фронтенда. Консоль: `http://localhost:9001` (логин/пароль: minioadmin).
- **Task Processor**: Автоматическая обработка задач генерации мемов. Настройки воркера через `.env`:
  - `TASK_PROCESSOR_WORKERS=10`
  - `TASK_PROCESSOR_QUEUE_SIZE=100`
  - `TASK_PROCESSOR_POLL_INTERVAL=5s`

## Генерация мемов

### Текущая реализация (для отладки)

```bash
POST /api/v1/memes/generate
Content-Type: multipart/form-data

"is_public": true,
"prompt": "я купил компьютер за 1000000",
"style": "anime" (доступные стили можно изучить по эндпоинту GET /api/v1/memes/styles)

```

мем создаётся со статусом `pending`, если нейронная сеть начала его обработку, статус изменяется на `started`

### Асинхронная генерация

1. Мем создаётся со статусом `pending` и возвращается `task_id`.
2. Worker берёт задачу из очереди и отправляет `prompt` в нейросеть.
3. Получает изображение и загружает в MinIO.
4. Обновляет статус на `completed`, появляется `image_url`.
5. Для проверки статуса используйте `/memes/{id}/status` или можно посмотреть значение поля в самому мем, получив его также через другой роутер.

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
