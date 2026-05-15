BIN_NAME = local-persist
LDFLAGS  = -s -w

.PHONY: test build binaries clean run docker

test:
	go vet ./...
	go test -v ./...

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME) ./cmd/local-persist

binaries: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-amd64 ./cmd/local-persist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-arm64 ./cmd/local-persist

clean:
	rm -rf bin/

run:
	sudo go run ./cmd/local-persist

docker:
	docker build -t $(BIN_NAME) .
