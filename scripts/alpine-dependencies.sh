#!/usr/bin/env sh

apk add fd just libxml2-utils shellcheck shfmt yamlfmt

wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.5.0
mv ./bin/golangci-lint /usr/bin/golangci-lint
