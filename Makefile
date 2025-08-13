MAIN := ./cmd/app/main.go
BIN_DIR := bin
TARGET := social_network
BIN   := $(BIN_DIR)/$(TARGET)
TESTS = ./internal/repo/users/my_sql/ ./internal/repo/posts/postmongo/ ./internal/handlers/users/ ./internal/handlers/posts/
POSTS_HANDLER_SRC := internal/handlers/posts/posts_handler.go
USERS_HANDLER_SRC := internal/handlers/users/users_handler.go
POSTS_HANDLER_MOCKS := internal/handlers/posts/posts_handler_mocks.go
USERS_HANDLER_MOCKS := internal/handlers/users/users_handler_mocks.go

.PHONY: start build run up down lint tests cover

start: up
	docker compose up -d --build

stop:
	docker compose down

build:
	@go build -v -o $(BIN) $(MAIN)

run:
	go run $(MAIN)

lint:
	golangci-lint run ./...

tests:
	go test $(TESTS)

cover:
	go test $(TESTS) -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
	rm cover.out

install_mocks:
	go install go.uber.org/mock/mockgen@v0.5.2

generate_mocks: install_mocks
	mockgen -source=$(POSTS_HANDLER_SRC) -destination=$(POSTS_HANDLER_MOCKS) -package=posts
	mockgen -source=$(USERS_HANDLER_SRC) -destination=$(USERS_HANDLER_MOCKS) -package=users