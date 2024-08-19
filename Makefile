CONFIG_PATH ?= ./config/.env
SERVER_PORT ?= 8082

# Don't forget to set POSTGRESQL_URL with your credentials
POSTGRESQL_URL ?='postgres://app:p%40ssw0rd@localhost:5432/reframed_dev?sslmode=disable'

.PHONY: migrate migrate-down run-server stop-server build test cover

setup: migrate

# Run migrations only if not already applied
migrate:
	@echo "Checking if postgresql-client is installed..."
	@if ! which psql > /dev/null 2>&1; then \
		echo "postgresql-client not found. Installing..."; \
		if [ "$$(uname)" = "Darwin" ]; then \
			echo "Detected macOS. Installing via Homebrew..."; \
			brew install postgresql; \
		elif [ "$$(uname)" = "Linux" ]; then \
			echo "Detected Linux. Installing via apt-get..."; \
			sudo apt-get update && sudo apt-get install -y postgresql-client; \
		else \
			echo "Unsupported OS. Please install postgresql-client manually."; \
			exit 1; \
		fi \
	else \
		echo "postgresql-client is already installed."; \
	fi

	@echo "Checking if migrations are needed..."
		@if psql $(POSTGRESQL_URL) -c "SELECT 1 FROM pg_tables WHERE tablename = 'tasks';" | grep -q 1; then \
			echo "Migrations are not needed."; \
		else \
			echo "Running migrations..."; \
			migrate -database $(POSTGRESQL_URL) -path migrations up; \
			echo "Migrations completed."; \
		fi

# Rollback migrations
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -database $(POSTGRESQL_URL) -path migrations down
	@echo "Migrations rolled back."

# Run server
run-server: stop-server
	@echo "Running the server..."
	@CONFIG_PATH=$(CONFIG_PATH) go run github.com/rshelekhov/reframed/cmd/reframed &
	@sleep 5 # Wait for the server to start
	@echo "Server is running with PID $$(lsof -t -i :$(SERVER_PORT))."

# Stop server
stop-server:
	@echo "Stopping the server..."
	@PID=$$(lsof -t -i :$(SERVER_PORT)); \
    	if [ -n "$$PID" ]; then \
    		kill $$PID; \
    		echo "Server stopped."; \
    	else \
    		echo "No server is running on port $(SERVER_PORT)."; \
    	fi

build: setup
	go build -v ./cmd/reframed

test: setup run-server
	@echo "Running tests..."
	go test ./api_tests
	@echo "Tests completed."

cover: setup run-server
	@echo "Running tests with coverage..."
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out
	@echo "Tests with coverage completed."

.DEFAULT_GOAL := build