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

### Мемы

- `GET /api/v1/memes` - Получить все мемы (публичный)
- `GET /api/v1/memes/:id` - Получить мем по ID (публичный)
- `POST /api/v1/memes/generate` - Сгенерировать мем (требует авторизации)
- `GET /api/v1/memes/my` - Получить свои мемы (требует авторизации)
- `DELETE /api/v1/memes/:id` - Удалить свой мем (требует авторизации)

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
MINIO_ROOT_PASSWORD=minioadmin
MINIO_ENDPOINT=minio:9000
MINIO_PUBLIC_URL=http://localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=memes
```

## Особенности

- **Авторизация**: Можно входить как по username, так и по email
- **JWT**: Access (1 час) + Refresh (7 дней) токены в HTTP-only cookies
- **Пароли**: Хешируются через Argon2
- **UUID**: Используются для всех ID
- **GORM**: Auto-миграции БД при старте
- **MinIO**: S3-совместимое хранилище для изображений мемов
- **Clean Architecture**: Разделение на слои handlers → services → repository

## Генерация мемов

### Текущая реализация (для отладки)

```bash
POST /api/v1/memes/generate
Content-Type: multipart/form-data

prompt: "ваш текст промпта"
image: файл изображения (опционально)
```

- **С файлом**: мем создаётся сразу со статусом `completed`
- **Без файла**: мем создаётся со статусом `pending` (для будущей нейронки)

### Будущая интеграция с нейронкой

Архитектура готова к интеграции:

1. Мем создаётся со статусом `pending`
2. Worker берёт задачу из очереди
3. Отправляет `prompt` в нейросеть
4. Получает изображение и загружает в MinIO
5. Обновляет статус на `completed`

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
make dev-db      # Запустить только PostgreSQL
make swagger     # Обновить документацию
make clean       # Очистить всё
```

## Деплой на сервер

Настроен автоматический деплой при push в `main` ветку через GitHub Actions.

Подробная инструкция по настройке: [.github/DEPLOY.md](.github/DEPLOY.md)

**Кратко:**

1. Добавьте GitHub Secrets: `SERVER_HOST`, `SERVER_USER`, `SSH_PRIVATE_KEY`
2. Подготовьте сервер (Docker, Git)
3. Push в `main` → автоматический деплой ✅
