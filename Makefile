.PHONY: test tests
tests: test
test:
	go test -race -timeout=60s -count 1 -v ./...

.PHONY: test_integration tests_integration
tests_integration: test_integration
test_integration: test_integration
	go test -race -timeout=60s -count 1 -v ./... -tags=integration

.PHONY: test_all tests_all
tests_all: test_all
test_all: test test_integration

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...
