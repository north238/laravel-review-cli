.PHONY: deps build test vet fmt shell

deps:
	docker compose run --rm dev go get github.com/spf13/cobra
	docker compose run --rm dev go get golang.org/x/sync/errgroup
	docker compose run --rm dev go mod tidy

build:
	docker compose run --rm dev go build -o bin/lrv ./cmd/lrv/

test:
	docker compose run --rm dev go test ./...

vet:
	docker compose run --rm dev go vet ./...

fmt:
	docker compose run --rm dev gofmt -l .

shell:
	docker compose run --rm dev sh
