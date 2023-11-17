GO_LDFLAGS := -ldflags="-s -w"
BINARY := kubedog
COVER_FILE := coverage.txt


all: generate check-dirty-repo build

generate: download
	go generate kubedog.go

build: test
	GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) -o ${BINARY} ./

test: fmt vet
	go test -race -timeout=300s -tags test -coverprofile=${COVER_FILE} ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: download
download:
	go mod download

.PHONY: lint
lint:
	@echo "golangci-lint"
	golangci-lint run ./...

.PHONY: cover
cover:
	@$(MAKE) test
	@go tool cover -html=${COVER_FILE}

.PHONY: check-dirty-repo
check-dirty-repo:
	@git diff --quiet HEAD || (\
	echo "Untracked files in git repo: " && \
	git status --short && \
	echo "- If 'docs/syntax.md' is up there, try running 'make generate' and commit the generated documentation" && \
	echo "- If 'go.mod' is up there, try running 'go mod tidy' and commit the changes" && \false)

.PHONY: clean
clean:
	@rm -f ${BINARY}
	@rm -f ${COVER_FILE}

