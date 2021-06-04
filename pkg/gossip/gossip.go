package gossip

import (
	"fmt"
	"net"
	"sync"
	"time"

	"gossip/pkg/membership"
)

type Gossip struct {
	L       *sync.Mutex
	Address *net.UDPAddr
	Members *membership.Membership
}

func GossipInit(nodeAddress *string, seedNodes []string) (*Gossip, error) {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *nodeAddress, 8000))
	if err != nil {
		return nil, err
	}

	members := make(map[string]time.Time)

	for _, node := range seedNodes {
		address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", node, 8000))
		if err != nil {
			return nil, err
		}
		members[address.String()] = time.Now()
	}

	return &Gossip{
		L:       &sync.Mutex{},
		Address: address,
		Members: &membership.Membership{Members: members},
	}, nil

}
