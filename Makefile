.PHONY: all
all: prerequisite prepare test testsum coverage check

.PHONY: prerequisite
prerequisite:
	@go install golang.org/x/tools/cmd/stringer@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/lint/golint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/client9/misspell/cmd/misspell@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest	
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@go install gotest.tools/gotestsum@latest

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