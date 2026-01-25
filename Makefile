.PHONY: test tests
tests: test
test:
	go test -race -timeout=60s -count 1 ./...

.PHONY: fmt
fmt:
	go tool goimports -local "github.com/nkiryanov/go-metrics" -w .

.PHONY: lint
lint:
	go tool staticcheck ./...
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build:
	cd cmd/server && go build .
	cd cmd/agent && go build .
