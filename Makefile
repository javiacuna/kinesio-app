.PHONY: fmt test lint run

fmt:
	go fmt ./...

test:
	go test ./...

run:
	go run ./cmd/api
