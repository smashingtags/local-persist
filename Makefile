BIN_NAME = local-persist
LDFLAGS  = -s -w

.PHONY: test build binaries clean run

test:
	GO_ENV=test go test -v ./...

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME) .

binaries: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-arm64 .

clean:
	rm -rf bin/

run:
	sudo go run main.go driver.go

docker:
	docker build -t $(BIN_NAME) .
