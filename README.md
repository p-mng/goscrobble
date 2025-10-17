# goscrobble 🎧💿

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/p-mng/goscrobble/go.yml) ![GitHub Tag](https://img.shields.io/github/v/tag/p-mng/goscrobble) ![AUR Version](https://img.shields.io/aur/version/goscrobble) [![Go Report Card](https://goreportcard.com/badge/github.com/p-mng/goscrobble)](https://goreportcard.com/report/github.com/p-mng/goscrobble)

## Description

`goscrobble` is a simple, cross-platform music scrobbler daemon using [MPRIS](https://specifications.freedesktop.org/mpris-spec/latest/)/[D-Bus](https://www.freedesktop.org/wiki/Software/dbus/) on Linux and [ungive/media-control](https://github.com/ungive/media-control) on macOS.

> [!WARNING]
> `goscrobble` is still beta software. Features may break without warning (especially on macOS), scrobbling may be unreliable, and the config file format is subject to change. Use at your own risk, as data loss is possible.

## Features

- **simple**: human-readable config file using [TOML](https://toml.io/en/)
- **lightweight**: command line interface, no GUI, minimal dependencies, fewer lines of code than the alternatives listed below
- **privacy-friendly**: no external services required, everything stays on your device (unless you use [last.fm](https://www.last.fm/))
- **cross-platform**: currently works on Linux (MPRIS) and macOS (media-control)
- **multi-player support**: supports scrobbling from multiple players (e.g. YouTube Music and Spotify) at the same time (only supported on Linux)

## Configuration

The following configuration file (created automatically in `~/.config/goscrobble/config.toml`) can be used to scrobble to last.fm and a local file:

```toml
# track position update frequency in seconds
poll_rate = 2
# minimum playback duration in seconds
min_playback_duration = 240
# minimum playback percentage
min_playback_percent = 50
# send a desktop notification when a scrobble is saved
notify_on_scrobble = false
# send a desktop notification when a scrobble cannot be saved
notify_on_error = true
# player blacklist
blacklist = ["chromium", "firefox"]

# regex match/replace
[[regexes]]
match = ' \([0-9]+ Remaster(ed)?\)'
replace = ""
artist = false
album = true
track = true

[[regexes]]
match = " - Radio Edit"
replace = " (Radio Edit)"
track = true

# last.fm configuration
[lastfm]
key = "<API key>"
secret = "<shared secret>"
session_key = "<session key (automatically generated using `goscrobble auth`)>"

# local file configuration (deprecated, use csv instead)
# [file]
# filename = "<file to write scrobbles to>"

# local CSV file configuration
[csv]
filename = "<file to write scrobbles to>"
```

You can blacklist one or multiple players using Go [regular expressions](https://gobyexample.com/regular-expressions). Players are identified by their D-Bus service name on Linux or the bundle identifier on macOS. The example above will block `org.mpris.MediaPlayer2.chromium.instance10670` and `org.mpris.MediaPlayer2.firefox.instance_1_84` on Linux and `org.mozilla.firefox` on macOS.

## Installation

### Linux, macOS

Install the newest version using the `go` toolchain:

```shell
go install github.com/p-mng/goscrobble@latest
```

Replace `@latest` with `@dev` for the latest development build.

On macOS, [media-control](https://github.com/ungive/media-control) and [julienXX/terminal-notifier](https://github.com/julienXX/terminal-notifier) are required:

```shell
brew install media-control terminal-notifier
```

### Arch Linux

Manually install the newest version from the [Arch User Repository](https://aur.archlinux.org/):

```shell
git clone https://aur.archlinux.org/goscrobble.git
cd goscrobble
makepkg -crsi
```

Or use an AUR helper like [paru](https://github.com/Morganamilo/paru):

```shell
paru -S goscrobble
```

### systemd user service

After creating the config file, start the systemd user service using `systemctl --user enable --now goscrobble.service`. If you installed the package using `go` directly from git, you might need to update the `.service` file with your correct binary location (likely `~/go/bin/goscrobble`) and copy the service file to `~/.config/systemd/user`.

## Connect last.fm account

1. [Create an API account](https://www.last.fm/api/account/create). Description, callback URL, and application homepage are not required.
2. Create the `lastfm` section in your config file and enter the [newly generated](https://www.last.fm/api/accounts) API key and shared secret.
3. Run `goscrobble auth`, and authenticate the application in your browser.
4. Confirm the following prompt with `y`, the session key will be automatically written to your config file.

## Known issues

### Double scrobbles when using tidal-hifi

tidal-hifi exposes two MPRIS media players (`tidal-hifi` and `chromium`). Right now, you should add `tidal-hifi` to your blacklist, as the playlist name is incorrectly reported as the album name (see [tidal-hifi issue 505](https://github.com/Mastermindzh/tidal-hifi/issues/505)).

## Similar projects

- [mariusor/mpris-scrobbler](https://github.com/mariusor/mpris-scrobbler): MPRIS scrobbler written in C
- [InputUsername/rescrobbled](https://github.com/InputUsername/rescrobbled): MPRIS scrobbler written in Rust
- [hrfee/go-scrobble](https://github.com/hrfee/go-scrobble): "ugly last.fm scrobbler" written in Go
- [web-scrobbler/web-scrobbler](https://github.com/web-scrobbler/web-scrobbler): Browser scrobbler written in TypeScript

I found all of the above to have different issues (e.g. [pausing breaks scrobbling](https://github.com/mariusor/mpris-scrobbler/issues/56) or [updates to the page layout preventing track detection](https://github.com/web-scrobbler/web-scrobbler/issues/4849)), so I decided to write my own scrobbler.
