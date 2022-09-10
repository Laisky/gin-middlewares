install:
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

lint:
	goimports -local github.com/Laisky/go-middlewares -w .
	go mod tidy
	go vet
	gofmt -s -w .
	govulncheck ./...
	golangci-lint run -c .golangci.lint.yml
