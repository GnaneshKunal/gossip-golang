package packet

import "gossip/pkg/message"

type Packet struct {
	Ip  string
	Msg *message.Message
}
