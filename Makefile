.PHONY: help build run dev-db dev-up dev-down swagger clean deploy deploy-setup

help:
	@echo "Доступные команды:"
	@echo "  build        - Собрать приложение"
	@echo "  run          - Запустить приложение локально"
	@echo "  dev-db       - Запустить только PostgreSQL для разработки"
	@echo "  dev-up       - Запустить всё в Docker для разработки"
	@echo "  dev-down     - Остановить Docker контейнеры"
	@echo "  swagger      - Обновить Swagger документацию"
	@echo "  clean        - Очистить bin/ и остановить контейнеры"
	@echo "  deploy-setup - Первоначальная настройка сервера"
	@echo "  deploy       - Деплой приложения на сервер"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

dev-db:
	docker-compose -f docker-compose.dev.yml up -d

dev-up:
	docker-compose up --build

dev-down:
	docker-compose down
	docker-compose -f docker-compose.dev.yml down

swagger:
	swag init -g cmd/server/main.go -o docs

clean:
	rm -rf bin/
	docker-compose down -v
	docker-compose -f docker-compose.dev.yml down -v

deploy-setup:
	@if [ -z "$(DEPLOY_HOST)" ]; then \
		echo "Error: DEPLOY_HOST is not set"; \
		echo "Usage: make deploy-setup DEPLOY_HOST=your-server.com"; \
		exit 1; \
	fi
	DEPLOY_HOST=$(DEPLOY_HOST) ./scripts/deploy.sh setup

deploy:
	@if [ -z "$(DEPLOY_HOST)" ]; then \
		echo "Error: DEPLOY_HOST is not set"; \
		echo "Usage: make deploy DEPLOY_HOST=your-server.com"; \
		exit 1; \
	fi
	DEPLOY_HOST=$(DEPLOY_HOST) ./scripts/deploy.sh deploy