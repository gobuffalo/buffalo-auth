GO_BIN ?= go

test:
	$(GO_BIN) test -failfast -short -cover ./...
	$(GO_BIN) mod tidy -v

cov:
	$(GO_BIN) test -short -coverprofile cover.out ./...
	$(GO_BIN) tool cover -html cover.out
	$(GO_BIN) mod tidy -v

install:
	$(GO_BIN) install -v .


build:
	$(GO_BIN) build -v .
	$(GO_BIN) mod tidy
