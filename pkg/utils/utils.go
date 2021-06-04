package utils

import (
	"encoding/json"
	"log"
	"net"

	"gossip/pkg/message"
)

func SendMessage(conn *net.UDPConn, dst *net.UDPAddr, msg *message.Message) (int, error) {

	sendMessageInternal := func(conn *net.UDPConn, dst *net.UDPAddr, msg *message.Message) (int, error) {

		encoded, err := json.Marshal(msg)
		if err != nil {
			return 0, err
		}

		return conn.WriteToUDP(encoded, dst)
	}

	count, err := sendMessageInternal(conn, dst, msg)
	if err != nil {
		return 0, &message.MsgError{
			Cause: err,
		}
	}

	return count, nil
}

func SendMessageRaw(conn *net.UDPConn, dstAddress string, msg *message.Message) (int, error) {
	dst, err := net.ResolveUDPAddr("udp", dstAddress)
	if err != nil {
		log.Println("Error resolving", err, dstAddress)
		return 0, nil
	}
	return SendMessage(conn, dst, msg)
}

func ReceiveMessage(conn *net.UDPConn) (*message.Message, *net.UDPAddr, error) {
	bytes := make([]byte, 2048)

	count, addr, err := conn.ReadFromUDP(bytes)
	if err != nil {
		return nil, nil, &message.MsgError{
			Cause: err,
		}
	}

	var msg message.Message
	err = json.Unmarshal(bytes[:count], &msg)
	if err != nil {
		return nil, nil, &message.MsgError{
			Cause: err,
		}
	}

	return &msg, addr, nil
}
