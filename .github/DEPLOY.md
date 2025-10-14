# Настройка автодеплоя

## GitHub Secrets

Для работы автодеплоя нужно добавить следующие секреты в Settings → Secrets and variables → Actions:

### SERVER_HOST

IP адрес или домен вашего сервера

Пример: 123.45.67.89 или server.example.com

### SERVER_USER

Имя пользователя для SSH подключения

Пример: root или projects

### SSH_PRIVATE_KEY

Приватный SSH ключ для подключения к серверу

Чтобы создать новый ключ:

ssh-keygen -t ed25519 -C "github-actions"

Скопируйте приватный ключ (~/.ssh/id_ed25519):

cat ~/.ssh/id_ed25519

И добавьте публичный ключ на сервер (~/.ssh/id_ed25519.pub):

ssh-copy-id -i ~/.ssh/id_ed25519.pub user@server

## Подготовка сервера

На сервере должно быть установлено:

- Docker
- Docker Compose
- Git

### Первичная настройка

1. Создайте .env файл на сервере:

ssh твой-сервер
cd $HOME/memology-backend
nano .env

Добавьте следующие переменные:

SERVER_PORT=8080
SERVER_HOST=0.0.0.0

DB*HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=ваш*сильный_пароль
DB_NAME=memology
DB_SSLMODE=disable

JWT*SECRET=ваш_jwt*секрет*минимум_32*символа
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=168h

Важно: Замените ваш*сильный*пароль и ваш*jwt*секрет на реальные значения!

2. Деплой запускается автоматически при push в main - репозиторий клонируется автоматически, если его нет.

## Как работает деплой

1. При push в main запускается GitHub Action
2. Подключается к серверу по SSH
3. Клонирует репозиторий (если его нет) или обновляет код
4. Собирает Docker образ из Dockerfile
5. Перезапускает контейнеры
6. Показывает последние 30 строк логов

## Проверка логов

Посмотреть логи деплоя можно в разделе Actions на GitHub.

На сервере логи можно посмотреть:

cd $HOME/memology-backend
docker logs memology_app --tail 50
docker logs memology_postgres --tail 50
docker compose logs -f

Последняя команда показывает все логи в реальном времени.

## Ручной перезапуск на сервере

Если нужно перезапустить вручную:

cd $HOME/memology-backend
docker compose down
docker compose build
docker compose up -d
docker logs memology_app --tail 30
