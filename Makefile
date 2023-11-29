GOLANGCI-LINT-VRS := "1.55.2"

.PHONY: all clean test lint

bin/jote:
	@CGO_ENABLED=0 go build -a -ldflags "-w" -v -o bin/jote cmd/jote/main.go

lint:
	@[ "$$(golangci-lint --version | awk '{print $$4}')" = "${GOLANGCI-LINT-VRS}" ] || \
	    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $$(go env GOPATH)/bin v${GOLANGCI-LINT-VRS}
	@golangci-lint run -c .golangci.yaml

test: lint
	@go vet ./...
	@go test -v -coverprofile coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
