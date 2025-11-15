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
    go test -coverprofile=cover.out -v ./...
    echo 'run `go tool cover -html cover.out` to view test coverage'

build:
    go build -v ./...
