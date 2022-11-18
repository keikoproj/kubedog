GO_LDFLAGS := -ldflags="-s -w"
BINARY := kubedog

all: style vet lint generate check-dirty-repo test build

.PHONY: generate
generate:
	go generate kubedog.go

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) $(BUILDARGS) -o ${BINARY} ./

.PHONY: test
test:
	go test -v -race -timeout=300s -tags test -coverprofile=coverage.txt ./...

.PHONY: style
style:
	gofmt -s -d -w .

.PHONY: vet
vet:
	go vet $(VETARGS) ./...

.PHONY: lint
lint:
	@echo "golangci-lint"
	golangci-lint run ./...

.PHONY: cover
cover:
	@$(MAKE) test TESTARGS="-tags test -coverprofile=coverage.txt"
	@go tool cover -html=coverage.txt

.PHONY: check-dirty-repo
check-dirty-repo:
	@git diff --quiet HEAD || (echo 'Untracked files in git repo: ' && git status --short && false)

.PHONY: clean
clean:
	@rm -f ${BINARY}
	@rm -f coverage.txt

