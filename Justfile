format:
    go fmt
    go mod tidy
    fd --hidden --extension yml --extension yaml --exec-batch yamlfmt

lint:
    golangci-lint run

test:
    go test -v ./...

build:
    go build -v ./...
