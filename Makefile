.PHONY: build
build:
	go build -v ./cmd/remedi

.PHONY: test
	go test -v -race -timeout 30s ./...

.PHONY: cover
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.DEFAULT_GOAL := build