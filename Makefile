BINARY_NAME=vault
BUILD_DIR=bin
APP_DIR=./cmd/vault

.PHONY: all build run clean test tidy

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(APP_DIR)

run:
	@echo "Running $(BINARY_NAME)..."
	go run $(APP_DIR)

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

tidy:
	@echo "Tidying modules..."
	go mod tidy
