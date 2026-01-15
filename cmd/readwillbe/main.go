package main

import (
	"encoding/gob"

	"readwillbe/views"
)

func init() {
	gob.Register(views.ManualReading{})
	gob.Register([]views.ManualReading{})
}

const SessionKey = "session"
const UserKey = "session-user"
const SessionUserIDKey = "userid"
const SessionLastSeenKey = "last_seen"
const SessionRefreshInterval = 3600

func main() {
	Execute()
}
