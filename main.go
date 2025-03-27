package main

import (
	"log"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Printf("failed to connect to session bus: %v", err)
		return
	}
	defer func(conn *dbus.Conn) {
		if err := conn.Close(); err != nil {
			log.Printf("error closing dbus connection: %v", err)
		}
	}(conn)

	config, err := ReadConfig()
	if err != nil {
		log.Printf("error reading config: %v", err)
		return
	}

	RunMainLoop(conn, config)
}
