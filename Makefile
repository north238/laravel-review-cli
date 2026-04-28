.PHONY: build test vet fmt

build:
	go build -o bin/lrv ./cmd/lrv/

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .
