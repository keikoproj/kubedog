# Default Go linker flags.
GO_LDFLAGS := -ldflags="-s -w"

# Step Functions
BINARY := kubedog

.PHONY: all
all: clean test build style vet lint

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) $(BUILDARGS) -o ${BINARY} ./

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: test
test:
	go test -v -race -timeout=300s -tags test -coverprofile=coverage.out ./...

.PHONY: style
style:
	gofmt -s -d -w .

.PHONY: vet
vet:
	go vet $(VETARGS) ./...

.PHONY: lint
lint:
	@echo "golint $(LINTARGS)"
	@for pkg in $(shell go list ./...) ; do \
		echo "golint $(LINTARGS) $$pkg" ; \
	done

.PHONY: goci
goci:
	@echo "golangci-lint"
	golangci-lint run ./...

.PHONY: cover
cover:
	@$(MAKE) test TESTARGS="-tags test -coverprofile=coverage.out"
	@go tool cover -html=coverage.out
	@rm -f coverage.out

.PHONY: clean
clean:
	@rm -rf ./build
