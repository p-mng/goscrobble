package main

type NotifierFunc func(replacesID uint32, summary, body string) (uint32, error)
