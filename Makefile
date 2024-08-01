.PHONY: test tests
tests: test
test:
	go test -race -timeout=60s -count 1 ./...


.PHONY: fmt
fmt:
	go fmt ./...
