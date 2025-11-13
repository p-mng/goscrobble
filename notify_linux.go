package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

func SendNotification(replacesID uint32, summary, body string) (uint32, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return 0, err
	}
	defer CloseDBus(conn)

	// https://specifications.freedesktop.org/notification/1.3/basic-design.html#id-1.3.6
	args := []any{"goscrobble", replacesID, "", summary, body, []string{}, map[string]dbus.Variant{}, int32(-1)}

	log.Debug().
		Interface("notification", args).
		Msg("sending desktop notification via dbus")

	var id uint32
	err = conn.
		Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications").
		Call("org.freedesktop.Notifications.Notify", 0, args...).
		Store(&id)

	log.Debug().
		Uint32("id", id).
		Msg("sent desktop notification")

	return id, err
}
