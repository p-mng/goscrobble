format:
    go fmt
    go mod tidy
    fd --hidden --extension yml --exec-batch yamlfmt
    fd --hidden --extension plist --exec xmllint --format --output {} {}

lint:
    golangci-lint run

test:
    go test -v ./...

build:
    go build -v ./...
