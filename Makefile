.PHONY: all
all: prerequisite prepare test testsum coverage check

.PHONY: prerequisite
prerequisite:
	@go install golang.org/x/tools/cmd/stringer@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install mvdan.cc/gofumpt@latest
	@go install gotest.tools/gotestsum@latest

.PHONY: update
update:
	@go get -u ./...
	@go mod tidy

.PHONY: prepare
prepare:
	gofumpt -l -w .

.PHONY: test
test: prepare
	go test ./...

.PHONY: testsum
testsum: prepare
	gotestsum

.PHONY: coverage
coverage:
	# Ignore (allow) packages without any tests
	go test ./... -coverprofile coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func coverage.out -o coverage.txt
	tail -1 coverage.txt

.PHONY: check
check: prepare
	golangci-lint run --fix