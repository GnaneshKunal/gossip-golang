package message

import (
	"encoding/json"
	"fmt"
	"time"

	"gossip/pkg/action"
	"gossip/pkg/membership"
	gt "gossip/pkg/time"
)

type Message struct {
	Id      int                   `json:"id"`
	Action  int                   `json:"action"`
	Time    time.Time             `json:"-"`
	Members membership.Membership `json:"members"`
}

type MessageAlias Message

type JSONMessage struct {
	MessageAlias
	Time gt.Time `json:"time"`
}

func NewJSONMessage(msg Message) JSONMessage {
	return JSONMessage{
		MessageAlias(msg),
		gt.Time{msg.Time},
	}
}

func (jm JSONMessage) Message() Message {
	msg := Message(jm.MessageAlias)
	msg.Time = msg.Time
	return msg
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

func (msg Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(NewJSONMessage(msg))
}

func (msg *Message) UnmarshalJSON(data []byte) error {
	var jsonMsg JSONMessage
	if err := json.Unmarshal(data, &jsonMsg); err != nil {
		return err
	}
	*msg = jsonMsg.Message()
	return nil
}
