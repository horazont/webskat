all: fmt vet test build

build:
	go build -trimpath ./cmd/webskat-server/webskat-server.go

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

.PHONY: fmt vet test
