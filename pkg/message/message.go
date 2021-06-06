package message

import (
	"fmt"
	"time"

	"gossip/pkg/action"
	"gossip/pkg/membership"
)

type Message struct {
	Id      int
	Action  int
	Time    time.Time
	Members membership.Membership
}

type JSONMessage struct {
	Id      int                   `json:"id"`
	Action  int                   `json:"action"`
	Time    int64                 `json:"time"`
	Members membership.Membership `json:"members"`
}

func NewJSONMessage(msg Message) JSONMessage {
	return JSONMessage{
		msg.Id,
		msg.Action,
		msg.Time.Unix(),
		msg.Members,
	}
}

func (msg JSONMessage) Message() Message {
	return Message{
		msg.Id,
		msg.Action,
		time.Unix(msg.Time, 0),
		msg.Members,
	}
}

func (msg Message) String() string {
	tm := msg.Time
	loc, _ := time.LoadLocation("Local")

	return fmt.Sprintf("{Id: %d, Action: %s, Time: %s}", msg.Id, action.ActionToString(msg.Action), tm.In(loc))
}

func NewMessage(action int) *Message {
	return &Message{
		Id:      int(time.Now().Unix()),
		Action:  action,
		Time:    time.Now(),
		Members: membership.Membership{Members: make(map[string]time.Time)},
	}
}
