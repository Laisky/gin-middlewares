install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0
	go get golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/protobuf/protoc-gen-go@v1.3.2
	go get -u github.com/go-bindata/go-bindata

lint:
	go mod tidy
	gofmt -s -w .
	golangci-lint run -E golint,depguard,gocognit,goconst,gofmt,misspell
