package action

import "log"

const (
	PING int = iota + 1
	PONG
	MEMBERSHIP
)

func ActionToString(action int) string {
	switch action {
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	default:
		log.Println("Unknown action")
		return ""
	}
}
