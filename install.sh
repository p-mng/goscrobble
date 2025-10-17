#!/usr/bin/env bash

GOPROXY=${GOPROXY:-"direct"}
GOSCROBBLE_VERSION=${GOSCROBBLE_VERSION:-"v0.5.0"}
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

GOSCROBBLE_PATH=$(which goscrobble)
GOSCROBBLE_PATH=${GOSCROBBLE_PATH:-"$HOME/go/bin/goscrobble"}

set -e

if [[ "$(uname -s)" = "Darwin" ]]; then
	echo "installing macOS launch daemon"

	mkdir -p "$HOME/Library/LaunchAgents"
	curl "$url_prefix/io.github.p-mng.goscrobble.plist" | envsubst >"$HOME/Library/LaunchAgents/io.github.p-mng.goscrobble.plist"
elif [[ "$(uname -s)" = "Linux" ]]; then
	echo "installing systemd user service"

	mkdir -p "$HOME/.config/systemd"
	curl "$url_prefix/goscrobble.service" | envsubst >"$HOME/.config/systemd/goscrobble.service"
else
	echo "unsupported platform: $(uname -s)"
	echo "please create a daemon manually according to your system's manual"
fi
