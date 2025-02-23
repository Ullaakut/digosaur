#Cleanup

fmt:
	@echo "==> Formatting source"
	@gofmt -s -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")
	@echo "==> Done"
.PHONY: fmt

#Test

test:
	@go test -cover -race ./...
.PHONY: test

#Lint

lint:
	@golangci-lint run --config=.golangci.yml ./...
.PHONY: lint

#Build

build:
	@goreleaser release --clean --snapshot
.PHONY: build

#Docs

openapi-gen:
	@go generate ./api/...
	@go run ./internal/cmd/openapi -o ./docs
.PHONY: openapi-gen

openapi-check:
	@echo "==> Checking OpenAPI Generation"
	@git diff --exit-code --quiet ./
.PHONY: openapi-check
