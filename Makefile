all: fmt vet test build

build:
	go build -trimpath ./cmd/skat-server/skat-server.go
	go build -trimpath ./cmd/skat-client/skat-client.go

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

.PHONY: fmt vet test
