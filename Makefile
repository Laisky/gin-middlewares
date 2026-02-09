install:
# 	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

lint:
	goimports -local github.com/Laisky/go-middlewares -w .
	go mod tidy
	go vet
	gofmt -s -w .
	golangci-lint run -c .golangci.lint.yml
	govulncheck ./...
