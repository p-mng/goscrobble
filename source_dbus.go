package main

import (
	"regexp"

	"github.com/godbus/dbus/v5"
)

type DBusSource struct {
	Conn *dbus.Conn
}

func (s DBusSource) Name() string {
	return "dbus"
}

func (s DBusSource) GetInfo(
	playerBlacklist []*regexp.Regexp,
	regexes []ParsedRegexReplace,
) (map[string]PlaybackStatus, error) {
	return nil, nil
}
