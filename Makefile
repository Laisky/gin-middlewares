.PHONY: install
install:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: lint
lint:
	go mod tidy
	govulncheck ./...
	goimports -local "github.com/Laisky/gin-middlewares" -w .
	gofmt -s -w .

	# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# golangci-lint run --timeout 3m -E golint,depguard,gocognit,goconst,gofmt,misspell,exportloopref,nilerr #,gosec,lll
	golangci-lint run -c .golangci.lint.yml
