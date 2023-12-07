build:
	go build -v ./cmd/remedi

.DEFAULT_GOAL := build
.PHONY: build