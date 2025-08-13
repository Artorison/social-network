MAIN := "./cmd/redditclone/main.go"
TESTS = ./internal/repo/users/my_sql/ ./internal/repo/posts/postmongo/ ./internal/handlers/users/ ./internal/handlers/posts/

start: up sleep run

run:
	go run $(MAIN)

run_jq:
	go run $(MAIN) | jq

lint:
	golangci-lint run ./...

up:
	docker compose up -d

down:
	docker compose down

tests:
	go test $(TESTS)

cover:
	go test $(TESTS) -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
	rm cover.out

sleep:
	sleep 10