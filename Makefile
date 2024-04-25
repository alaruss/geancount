.PHONY: all build fmt vet test cov

all: fmt vet test build

build: 
	go build -v -o bin/geancount

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

cov:
	go test ./... -cover -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html