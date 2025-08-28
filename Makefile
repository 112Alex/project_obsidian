.PHONY: build run test clean docker-build docker-run docker-compose-up docker-compose-down lint

# Переменные
BINARY_NAME=app
DOCKER_IMAGE=project_obsidian
DOCKER_TAG=latest

# Сборка приложения
build:
	go build -o $(BINARY_NAME) ./cmd/app

# Запуск приложения
run:
	go run ./cmd/app

# Запуск тестов
test:
	go test -v ./...

# Запуск тестов с покрытием
test-coverage:
	go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Очистка бинарных файлов
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Сборка Docker образа
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Запуск Docker контейнера
docker-run:
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Запуск с использованием Docker Compose
docker-compose-up:
	docker-compose up -d

# Остановка Docker Compose
docker-compose-down:
	docker-compose down

# Линтинг кода
lint:
	golangci-lint run

# Создание миграции
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Применение миграций
migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/obsidian?sslmode=disable" up

# Откат миграций
migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/obsidian?sslmode=disable" down

# Генерация документации
docs:
	swag init -g cmd/app/main.go -o docs

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build              - Сборка приложения"
	@echo "  run                - Запуск приложения"
	@echo "  test               - Запуск тестов"
	@echo "  test-coverage      - Запуск тестов c покрытием"
	@echo "  clean              - Очистка бинарных файлов"
	@echo "  docker-build       - Сборка Docker образа"
	@echo "  docker-run         - Запуск Docker контейнера"
	@echo "  docker-compose-up  - Запуск c использованием Docker Compose"
	@echo "  docker-compose-down - Остановка Docker Compose"
	@echo "  lint               - Линтинг кода"
	@echo "  migrate-create     - Создание миграции"
	@echo "  migrate-up         - Применение миграций"
	@echo "  migrate-down       - Откат миграций"
	@echo "  docs               - Генерация документации"
	@echo "  help               - Показать эту справку"