package message

import (
	"fmt"
	"strconv"
	"time"

	"gossip/pkg/action"
	"gossip/pkg/membership"
)

type Message struct {
	Id      int                   `json:"id"`
	Action  int                   `json:"action"`
	Time    string                `json:"time"`
	Members membership.Membership `json:"members"`
}

func (msg Message) String() string {
	timestamp, _ := time.Parse(time.RFC3339, msg.Time)
	tm := timestamp
	loc, _ := time.LoadLocation("Asia/Kolkata")

	return fmt.Sprintf("{Id: %d, Action: %s, Time: %s}", msg.Id, action.ActionToString(msg.Action), tm.In(loc))
}

type MsgError struct {
	Cause error
}

func NewMessage(action int) *Message {
	return &Message{
		Id:      int(time.Now().Unix()),
		Action:  action,
		Time:    strconv.FormatInt(time.Now().Unix(), 10),
		Members: membership.Membership{Members: make(map[string]time.Time)},
	}
}

func (msg *MsgError) Error() string {
	return fmt.Sprintf("Cause %s\n", msg.Cause)
}
