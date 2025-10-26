format:
    go fmt
    go mod tidy
    fd --hidden --extension yml --exec-batch yamlfmt
    fd --hidden --extension plist --exec xmllint --format --output {} {}
    fd --hidden --extension sh --exec-batch  shfmt --write

lint:
    golangci-lint run
    shellcheck scripts/alpine-dependencies.sh
    shellcheck scripts/install.sh

test:
    go test -v ./...

build:
    go build -v ./...
