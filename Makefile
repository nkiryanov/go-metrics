.PHONY: test tests
tests: test
test:
	go test -race -timeout=60s -count 1 -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONE: generate
generate:
	go generate ./...

.PHONE: build
build:
	cd cmd/server && go build .
	cd cmd/agent && go build .
