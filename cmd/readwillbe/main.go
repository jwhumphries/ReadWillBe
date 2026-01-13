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

func main() {
	Execute()
}
