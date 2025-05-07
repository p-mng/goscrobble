# goscrobble 🎧💿

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/p-mng/goscrobble/go.yml) ![GitHub Tag](https://img.shields.io/github/v/tag/p-mng/goscrobble) ![AUR Version](https://img.shields.io/aur/version/goscrobble)

## Description

`goscrobble` is a simple music scrobbler daemon for MPRIS-based music players (e.g. [Spotify](https://www.spotify.com/en/download/linux/) or [tidal-hifi](https://github.com/Mastermindzh/tidal-hifi)).

## Features

- **simple**: human-readable config file using [TOML](https://toml.io/en/)
- **lightweight**: command line interface, no GUI, minimal dependencies, fewer lines of code than the alternatives listed below
- **privacy-friendly**: no external services required, everything stays on your device (unless you use [last.fm](https://www.last.fm/))
- **multi-player support**: supports scrobbling from multiple players (e.g. YouTube Music and Spotify) at the same time

## Configuration

The following configuration file (created automatically in `~/.config/goscrobble/config.toml`) can be used to scrobble to last.fm and a local file:

```toml
# track position update frequency in seconds
poll_rate = 2
# minimum playback duration in seconds
min_playback_duration = 240
# minimum playback percentage
min_playback_percent = 50
# MPRIS player blacklist
blacklist = [ "chromium", "firefox" ]

# last.fm configuration
[lastfm]
key = "<API key>"
secret = "<shared secret>"
session_key = "<session key (automatically generated using goscrobble auth)>"

# local file configuration
[file]
filename = "<file to write scrobbles to>"
```

You can blacklist one or multiple players using [regular expressions](https://gobyexample.com/regular-expressions). The example above will block both `org.mpris.MediaPlayer2.chromium.instance10670` and `org.mpris.MediaPlayer2.firefox.instance_1_84`.

If you don't want to use one of the two supported providers (`lastfm` or `file`), just remove the section from the config file.

## Installation

Install using the `go` toolchain:

```
go install github.com/p-mng/goscrobble@latest
```

Install from the [Arch User Repository](https://aur.archlinux.org/):

```
git clone https://aur.archlinux.org/goscrobble.git
cd goscrobble
makepkg -crsi
```

After creating the config file, start the systemd user service using `systemctl --user enable --now goscrobble.service`. If you installed the package using `go` directly from git, you might need to update the `.service` file with your correct binary location (likely `~/go/bin/goscrobble`) and copy the service file to `~/.config/systemd/user`.

## Connect last.fm account

1. [Create an API account](https://www.last.fm/api/account/create). Description, callback URL, and application homepage are not required.
2. Create the `lastfm` section in your config file and enter the newly generated API key and shared secret.
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
