package main

import "github.com/godbus/dbus/v5"

// https://specifications.freedesktop.org/icon-naming-spec/latest/
const (
	IconSyncError    = "sync-error"
	IconSyncronizing = "sync-synchronizing"
)

// https://dbus.freedesktop.org/doc/dbus-specification.html
// https://specifications.freedesktop.org/notification-spec/1.3/
func SendNotification(conn *dbus.Conn, replacesID uint32, appIcon, summary, body string) (uint32, error) {
	args := []any{
		"goscrobble",
		replacesID,
		appIcon,
		summary,
		body,
		[]string{},
		map[string]dbus.Variant{},
		int32(-1),
	}

	var id uint32
	err := conn.
		Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications").
		Call("org.freedesktop.Notifications.Notify", 0, args...).
		Store(&id)
	return id, err
}
