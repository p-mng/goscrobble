# goscrobble 🎧💿

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/p-mng/goscrobble/go.yml) ![GitHub Tag](https://img.shields.io/github/v/tag/p-mng/goscrobble) ![AUR Version](https://img.shields.io/aur/version/goscrobble) [![Go Report Card](https://goreportcard.com/badge/github.com/p-mng/goscrobble)](https://goreportcard.com/report/github.com/p-mng/goscrobble)

## Description

A simple, cross-platform music scrobbler daemon. Inspired by audio software like PulseAudio and PipeWire, it can be configured to connect different *sources* and *sinks*. Scrobbles are sent from one or more sources to one or more sinks.

Supported sources:

- [D-Bus (MPRIS2)](https://specifications.freedesktop.org/mpris-spec/latest/): works best on **Linux**
- [media-control](https://github.com/ungive/media-control): works best on **macOS**

Supported sinks:

- **last.fm**: "the world's largest online music service"
- **CSV**: a local file with comma-separated data

> [!WARNING]
> This project is still beta software. Features may break without warning (especially on macOS), scrobbling may be unreliable, and the config file format is subject to change. Use at your own risk.
>
> This README refers to the `main` branch. To view the README for a specific version, check out the corresponding tagged commit.

## Features

- **simple**: human-readable config file using [TOML](https://toml.io/en/)
- **lightweight**: command line interface, no GUI, minimal dependencies, fewer lines of code than many of the alternatives listed below
- **privacy-friendly**: no internet connection or external services required
- **cross-platform**: currently works on Linux (MPRIS2/D-Bus) and macOS (media-control)
- **multi-player support**: supports scrobbling from multiple players (e.g., YouTube Music and Spotify) at the same time (only supported on Linux)

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
match = ' - [0-9]+ Remaster(ed)?'
replace = ''
artist = false
track = true
album = true

[[regexes]]
match = " - Radio Edit"
replace = " (Radio Edit)"
track = true

[sources.dbus]
# dbus address: if empty, connect to the session bus
address = ''

[sources.media-control]
# path to the 'media-control' binary
command = 'media-control'
# arguments, no need to change this
arguments = ['get', '--now']

[sinks.lastfm]
# last.fm API key
key = 'last.fm API key'
# last.fm API shared secret
secret = 'last.fm API secret'
# last.fm session key, automatically set by 'goscrobble lastfm-auth'
session_key = ''
# last.fm username, automatically set by 'goscrobble lastfm-auth'
username = ''

[sinks.csv]
# filename to write scrobbles to, defaults to $HOME/scrobbles.csv
filename = '/Users/patrick/scrobbles.csv'
```

You can blacklist one or more players using Go [regular expressions](https://gobyexample.com/regular-expressions). Players are identified by their D-Bus service name on Linux or the bundle identifier on macOS. The example above will block `org.mpris.MediaPlayer2.chromium.instance10670` and `org.mpris.MediaPlayer2.firefox.instance_1_84` on Linux and `org.mozilla.firefox` on macOS.

## Installation

### Linux, macOS (install script)

Using the provided install script:

```shell
curl https://raw.githubusercontent.com/p-mng/goscrobble/refs/heads/main/scripts/install.sh | bash
```

Set `GOSCROBBLE_VERSION` to a branch or tag name to install a specific version (e.g., `export GOSCROBBLE_VERSION=dev`).

On macOS, [media-control](https://github.com/ungive/media-control) and [julienXX/terminal-notifier](https://github.com/julienXX/terminal-notifier) are required:

```shell
brew install media-control terminal-notifier
```

### Arch Linux

Manually install the latest version from the [Arch User Repository](https://aur.archlinux.org/):

```shell
git clone https://aur.archlinux.org/goscrobble.git
cd goscrobble
makepkg -crsi
```

Or use an AUR helper like [paru](https://github.com/Morganamilo/paru):

```shell
paru -S goscrobble
```

## Connect last.fm account

1. [Create an API account](https://www.last.fm/api/account/create). Description, callback URL, and application homepage are not required.
2. Open the config file and insert the [newly generated API key and shared secret](https://www.last.fm/api/accounts).
3. Run `goscrobble lastfm-auth`, and authenticate the application in your browser.
4. Confirm the following prompt. The session key and last.fm username will be automatically written to your config file.

## Known issues

### Double scrobbles when using tidal-hifi

tidal-hifi exposes two MPRIS media players (`tidal-hifi` and `chromium`). Right now, you should add `tidal-hifi` to your blacklist, as the playlist name is incorrectly reported as the album name (see [tidal-hifi issue 505](https://github.com/Mastermindzh/tidal-hifi/issues/505)).

## TODOs

- Add more scrobbling sinks (e.g., Maloja, LibreFM, ListenBrainz, etc.)
- Add Microsoft Windows support
- Test more Linux distros, macOS versions, and music players
- Add unit/integration tests

## Similar projects

- [FoxxMD/multi-scrobbler](https://github.com/FoxxMD/multi-scrobbler): scrobbler for multiple sources and clients (written in TypeScript)
- [mariusor/mpris-scrobbler](https://github.com/mariusor/mpris-scrobbler): MPRIS scrobbler (written in C)
- [InputUsername/rescrobbled](https://github.com/InputUsername/rescrobbled): MPRIS scrobbler (written in Rust)
- [hrfee/go-scrobble](https://github.com/hrfee/go-scrobble): "ugly last.fm scrobbler" (written in Go)
- [web-scrobbler/web-scrobbler](https://github.com/web-scrobbler/web-scrobbler): scrobbling browser extension (written in TypeScript)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
