#!/usr/bin/env bash

export GOPROXY=${GOPROXY:-"direct"}
export GOSCROBBLE_VERSION=${GOSCROBBLE_VERSION:-"v0.6.0"}
REQUIRED_BINS=("go" "envsubst" "sed" "curl")

is_tag() {
	if [[ "$1" =~ ^v[0-9]\.[0-9]\.[0-9]$ ]]; then
		return 0
	else
		return 1
	fi
}

for bin in "${REQUIRED_BINS[@]}"; do
	if ! command -v "$bin" &>/dev/null; then
		echo "$bin is not installed"
		exit 10
	fi
done

echo "installing github.com/p-mng/goscrobble@$GOSCROBBLE_VERSION"
go install "github.com/p-mng/goscrobble@$GOSCROBBLE_VERSION"

if is_tag "$GOSCROBBLE_VERSION"; then
	echo "${GOSCROBBLE_VERSION} is a tag"
	url_prefix="https://raw.githubusercontent.com/p-mng/goscrobble/refs/tags/${GOSCROBBLE_VERSION}"
else
	echo "${GOSCROBBLE_VERSION} is a branch name"
	url_prefix="https://raw.githubusercontent.com/p-mng/goscrobble/refs/heads/${GOSCROBBLE_VERSION}"
fi

if [[ -z "$(which goscrobble)" ]]; then
	GOPATH=${GOPATH:-"$HOME/go"}

	echo "goscrobble not found in path, assuming installation at $GOPATH/goscrobble"
	GOSCROBBLE_PATH="$GOPATH/go/bin/goscrobble"
else
	echo "goscrobble installation found at $(which goscrobble)"
	GOSCROBBLE_PATH="$(which goscrobble)"
fi

export GOSCROBBLE_PATH

set -e

if [[ "$(uname -s)" = "Darwin" ]]; then
	echo "installing macOS launch daemon"

	mkdir -p "$HOME/Library/LaunchAgents"
	curl "$url_prefix/scripts/io.github.p-mng.goscrobble.plist" | envsubst >"$HOME/Library/LaunchAgents/io.github.p-mng.goscrobble.plist"
elif [[ "$(uname -s)" = "Linux" ]]; then
	echo "installing systemd user service"

	mkdir -p "$HOME/.config/systemd/user"
	curl "$url_prefix/scripts/goscrobble.service" | envsubst >"$HOME/.config/systemd/user/goscrobble.service"
else
	echo "unsupported platform: $(uname -s)"
	echo "please create a daemon manually according to your system's manual"
fi
