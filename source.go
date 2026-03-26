package main

type Source interface {
	Name() string
	GetInfo() (map[string]PlaybackStatus, error)
}
