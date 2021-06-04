package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"gossip/pkg/action"
	g "gossip/pkg/gossip"
	"gossip/pkg/message"
	"gossip/pkg/packet"
	"gossip/pkg/utils"
)

func main() {

	rand.Seed(time.Now().Unix())

	addressPtr := flag.String("address", ":8000", "Node address")

	var seedNodesArg string
	flag.StringVar(&seedNodesArg, "seeds", "", "seed nodes")

	flag.Parse()

	seedNodes := strings.Split(seedNodesArg, ",")

	gossip, err := g.GossipInit(addressPtr, seedNodes)
	if err != nil {
		log.Println(err)
		return
	}

	conn, err := net.ListenUDP("udp", gossip.Address)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Listening on ", gossip.Address)

	done := make(chan bool)
	processor := make(chan packet.Packet, 1000)
	sender := make(chan packet.Packet, 1000)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		memTicker := time.NewTicker(5 * time.Second)
		garbageTicker := time.NewTicker(5 * time.Second)
		for {
			select {

			case <-memTicker.C:

				msg := message.NewMessage(action.MEMBERSHIP)

				gossip.L.Lock()
				dst, _ := gossip.Members.Random()
				msg.Members = *gossip.Members
				gossip.L.Unlock()

				sender <- packet.Packet{
					Ip:  dst,
					Msg: msg,
				}
				gossip.L.Lock()
				log.Println("members: ", gossip.Members)
				gossip.L.Unlock()

			case packet := <-sender:

				_, err := utils.SendMessageRaw(conn, packet.Ip, packet.Msg)
				if err != nil {
					log.Println("Err: ", err)
					continue
				}

			case pack := <-processor:
				msg := pack.Msg
				gossip.L.Lock()
				gossip.Members.Touch(pack.Ip)
				gossip.L.Unlock()
				switch msg.Action {
				case action.PING:
					sender <- packet.Packet{
						Ip:  pack.Ip,
						Msg: message.NewMessage(action.PONG),
					}
				case action.PONG:
					// log.Println("Received PONG", packet.ip)
				case action.MEMBERSHIP:

					// remove self entry
					delete(msg.Members.Members, gossip.Address.String())

					gossip.L.Lock()
					gossip.Members.Merge(msg.Members)
					gossip.L.Unlock()
				}

			case <-ticker.C:
				// heartbeat
				for i := 0; i < 3; i++ {
					gossip.L.Lock()
					dst, _ := gossip.Members.Random()
					gossip.L.Unlock()
					sender <- packet.Packet{
						Ip:  dst,
						Msg: message.NewMessage(action.PING),
					}
				}

			case <-garbageTicker.C:
				gossip.L.Lock()
				for node, timestamp := range gossip.Members.Members {
					if time.Since(timestamp).Seconds() > 30 {
						log.Println(node, "DEAD")
						delete(gossip.Members.Members, node)
					}
				}
				gossip.L.Unlock()

			default:
				time.Sleep(100 * time.Millisecond)
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				msg, addr, err := utils.ReceiveMessage(conn)
				if err != nil {
					continue
				}
				processor <- packet.Packet{
					Ip:  addr.String(),
					Msg: msg,
				}
			}

		}
	}()

	<-done
}
