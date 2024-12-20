package main

import (
	"log"

	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalf("failed to connect to session bus: %v", err)
	}
	defer conn.Close()

	config, err := ReadConfig()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	RunMainLoop(conn, config)
}
