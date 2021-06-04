package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

type Message struct {
	Id      int        `json:"id"`
	Action  int        `json:"action"`
	Time    string     `json:"time"`
	Members Membership `json:"members"`
}

func (msg Message) String() string {
	timestamp, _ := time.Parse(time.RFC3339, msg.Time)
	tm := timestamp
	loc, _ := time.LoadLocation("Asia/Kolkata")

	return fmt.Sprintf("{Id: %d, Action: %s, Time: %s}", msg.Id, ActionToString(msg.Action), tm.In(loc))
}

type MsgError struct {
	cause error
}

func NewMessage(action int) *Message {
	return &Message{
		Id:      int(time.Now().Unix()),
		Action:  action,
		Time:    strconv.FormatInt(time.Now().Unix(), 10),
		Members: Membership{Members: make(map[string]time.Time)},
	}
}

func (msg *MsgError) Error() string {
	return fmt.Sprintf("Cause %s\n", msg.cause)
}

type Packet struct {
	ip  string
	msg *Message
}

func sendMessage(conn *net.UDPConn, dst *net.UDPAddr, msg *Message) (int, error) {

	sendMessageInternal := func(conn *net.UDPConn, dst *net.UDPAddr, msg *Message) (int, error) {

		encoded, err := json.Marshal(msg)
		if err != nil {
			return 0, err
		}

		return conn.WriteToUDP(encoded, dst)
	}

	count, err := sendMessageInternal(conn, dst, msg)
	if err != nil {
		return 0, &MsgError{
			cause: err,
		}
	}

	return count, nil
}

func receiveMessage(conn *net.UDPConn) (*Message, *net.UDPAddr, error) {
	bytes := make([]byte, 2048)

	count, addr, err := conn.ReadFromUDP(bytes)
	if err != nil {
		return nil, nil, &MsgError{
			cause: err,
		}
	}

	var msg Message
	err = json.Unmarshal(bytes[:count], &msg)
	if err != nil {
		return nil, nil, &MsgError{
			cause: err,
		}
	}

	return &msg, addr, nil
}

type Membership struct {
	Members map[string]time.Time
}

func (self *Membership) String() string {
	val := "{"
	for k, _ := range self.Members {
		val = val + k + ","
		// val = val + strconv.FormatInt(v.Unix(), 10) + ","
	}
	val = val + "}"
	return val
}

// func (self *Membership) MarshalJSON() ([]byte, error) {
// 	members := make(map[string]string)
// 	for member, timestamp := range self.Members {
// 		members[member] = strconv.FormatInt(timestamp.Unix(), 10)
// 	}

// 	return json.Marshal(members)
// }

func NewMembership() *Membership {
	return &Membership{
		Members: make(map[string]time.Time),
	}
}

func (self *Membership) touch(addr string) {
	if _, ok := self.Members[addr]; !ok {
		log.Println("Adding new node", addr)
	}

	self.Members[addr] = time.Now()
}

func (self *Membership) merge(members Membership) {
	for node, timestamp := range members.Members {

		duration := time.Since(timestamp)

		if duration.Seconds() > 30 {
			log.Println("Tried to add an expired node", node)
			continue
		}

		self.Members[node] = timestamp
	}
}

func (self *Membership) random() (string, time.Time) {
	r := rand.Intn(len(self.Members))
	count := 0

	for k, v := range self.Members {
		if count == r {
			return k, v
		}
		count += 1
	}

	// garbage
	return "", time.Now()
}

type Gossip struct {
	L         *sync.Mutex
	address   *net.UDPAddr
	seedNodes []*net.UDPAddr
	Members   *Membership
}

func GossipInit(nodeAddress *string, seedNodes []string) (*Gossip, error) {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *nodeAddress, 8000))
	if err != nil {
		return nil, err
	}

	var seedAddresses []*net.UDPAddr
	members := make(map[string]time.Time)

	for _, node := range seedNodes {
		address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", node, 8000))
		if err != nil {
			return nil, err
		}
		seedAddresses = append(seedAddresses, address)
		members[address.String()] = time.Now()
	}

	return &Gossip{
		L:         &sync.Mutex{},
		address:   address,
		seedNodes: seedAddresses,
		Members:   &Membership{Members: members},
	}, nil

}

func sendMessageRaw(conn *net.UDPConn, dstAddress string, msg *Message) (int, error) {
	dst, err := net.ResolveUDPAddr("udp", dstAddress)
	if err != nil {
		log.Println("Error resolving", err, dstAddress)
		return 0, nil
	}
	return sendMessage(conn, dst, msg)
}

func main() {

	rand.Seed(time.Now().Unix())

	addressPtr := flag.String("address", ":8000", "Node address")

	var seedNodesArg string
	flag.StringVar(&seedNodesArg, "seeds", "", "seed nodes")

	flag.Parse()

	seedNodes := strings.Split(seedNodesArg, ",")

	gossip, err := GossipInit(addressPtr, seedNodes)
	if err != nil {
		log.Println(err)
		return
	}

	conn, err := net.ListenUDP("udp", gossip.address)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Listening on ", gossip.address)

	done := make(chan bool)
	processor := make(chan Packet, 1000)
	sender := make(chan Packet, 1000)

	name, err := os.Hostname()
	log.Println("hostname: ", name)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		memTicker := time.NewTicker(5 * time.Second)
		garbageTicker := time.NewTicker(5 * time.Second)
		for {
			select {

			case <-memTicker.C:

				msg := NewMessage(MEMBERSHIP)

				gossip.L.Lock()
				dst, _ := gossip.Members.random()
				msg.Members = *gossip.Members
				gossip.L.Unlock()

				sender <- Packet{
					ip:  dst,
					msg: msg,
				}
				gossip.L.Lock()
				log.Println(gossip.Members)
				gossip.L.Unlock()

			case packet := <-sender:

				_, err := sendMessageRaw(conn, packet.ip, packet.msg)
				if err != nil {
					log.Println("Err: ", err)
					continue
				}

			case packet := <-processor:
				msg := packet.msg
				gossip.L.Lock()
				gossip.Members.touch(packet.ip)
				gossip.L.Unlock()
				switch msg.Action {
				case PING:
					sender <- Packet{
						ip:  packet.ip,
						msg: NewMessage(PONG),
					}
				case PONG:
					// log.Println("Received PONG", packet.ip)
				case MEMBERSHIP:
					// remove my entry
					delete(msg.Members.Members, gossip.address.String())

					gossip.L.Lock()
					gossip.Members.merge(msg.Members)
					gossip.L.Unlock()
				}

			case <-ticker.C:
				// heartbeat
				for i := 0; i < 3; i++ {
					gossip.L.Lock()
					dst, _ := gossip.Members.random()
					gossip.L.Unlock()
					sender <- Packet{
						ip:  dst,
						msg: NewMessage(PING),
					}
				}

			case <-garbageTicker.C:
				gossip.L.Lock()
				for node, timestamp := range gossip.Members.Members {
					if time.Since(timestamp).Seconds() > 30 {
						log.Println(node, "EXPIRED")
						delete(gossip.Members.Members, node)
					}
				}
				gossip.L.Unlock()

			default:
				time.Sleep(100 * time.Millisecond)
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				msg, addr, err := receiveMessage(conn)
				if err != nil {
					continue
				}
				processor <- Packet{
					ip:  addr.String(),
					msg: msg,
				}
			}

		}
	}()

	<-done
}
