# Настройка автодеплоя

## GitHub Secrets

Для работы автодеплоя нужно добавить следующие секреты в Settings → Secrets and variables → Actions:

### SERVER_HOST

IP адрес или домен вашего сервера

```
Пример: 123.45.67.89 или server.example.com
```

### SERVER_USER

Имя пользователя для SSH подключения

```
Пример: root или ubuntu
```

### SSH_PRIVATE_KEY

Приватный SSH ключ для подключения к серверу

Чтобы создать новый ключ:

```bash
ssh-keygen -t ed25519 -C "github-actions"
```

Скопируйте **приватный** ключ (`~/.ssh/id_ed25519`):

```bash
cat ~/.ssh/id_ed25519
```

И добавьте **публичный** ключ на сервер (`~/.ssh/id_ed25519.pub`):

```bash
ssh-copy-id -i ~/.ssh/id_ed25519.pub user@server
```

## Подготовка сервера

На сервере должен быть установлен:

- Docker
- Docker Compose
- Git

### Первичная настройка

1. Клонируйте репозиторий на сервер:

```bash
mkdir -p ~/projects
cd ~/projects
git clone https://github.com/lDizil/memology-backend.git
cd memology-backend
```

2. Создайте `.env` файл на сервере:

```bash
cp .env.example .env
nano .env
```

3. Запустите вручную первый раз:

```bash
docker compose up -d
```

После этого каждый push в `main` ветку будет автоматически деплоить изменения на сервер!

## Как работает деплой

1. При push в `main` запускается GitHub Action
2. Подключается к серверу по SSH
3. Обновляет код из git репозитория
4. Пересобирает Docker контейнеры
5. Перезапускает приложение
6. Проверяет что всё работает

## Проверка логов

Посмотреть логи деплоя можно в разделе Actions на GitHub.

На сервере логи можно посмотреть:

```bash
docker logs memology_app
docker logs memology_postgres
```
