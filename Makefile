.PHONY: help build run dev-db dev-up dev-down swagger clean

help:
	@echo "Доступные команды:"
	@echo "  build     - Собрать приложение"
	@echo "  run       - Запустить приложение локально"
	@echo "  dev-up    - Запустить всё в Docker для разработки"
	@echo "  dev-down  - Остановить Docker контейнеры"
	@echo "  swagger   - Обновить Swagger документацию"
	@echo "  clean     - Очистить bin/ и остановить контейнеры"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

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