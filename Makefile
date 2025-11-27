.PHONY: help build run test clean docker-build docker-run docker-stop dev

# Переменные
APP_NAME=fcm-push-service
DOCKER_IMAGE=$(APP_NAME):latest
MAIN_PATH=./cmd/server/main.go

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Собрать бинарник
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: bin/$(APP_NAME)"

run: ## Запустить сервис
	@echo "Starting $(APP_NAME)..."
	@go run $(MAIN_PATH)

dev: ## Запустить в режиме разработки (с hot reload)
	@./run-dev.sh

test: ## Запустить тесты
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Запустить линтер
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Форматировать код
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

tidy: ## Очистить зависимости
	@echo "Tidying dependencies..."
	@go mod tidy

clean: ## Очистить сгенерированные файлы
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Собрать Docker образ
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Запустить Docker контейнер
	@echo "Starting Docker container..."
	@docker-compose up -d
	@echo "Container started"

docker-stop: ## Остановить Docker контейнер
	@echo "Stopping Docker container..."
	@docker-compose down
	@echo "Container stopped"

docker-logs: ## Показать логи Docker контейнера
	@docker-compose logs -f

install-tools: ## Установить необходимые инструменты
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed"

# Для CI/CD
ci-test: ## Запустить тесты для CI
	@go test -v -race -coverprofile=coverage.out ./...

ci-build: ## Собрать для CI
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) $(MAIN_PATH)
