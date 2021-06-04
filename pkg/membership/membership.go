package membership

import (
	"log"
	"math/rand"
	"time"
)

type Membership struct {
	Members map[string]time.Time
}

func (self *Membership) String() string {
	val := "{"
	for k, _ := range self.Members {
		val = val + k + ","
	}
	val = val + "}"
	return val
}

func NewMembership() *Membership {
	return &Membership{
		Members: make(map[string]time.Time),
	}
}

func (self *Membership) Touch(addr string) {
	if _, ok := self.Members[addr]; !ok {
		log.Println("Adding new node", addr)
	}

	self.Members[addr] = time.Now()
}

func (self *Membership) Merge(members Membership) {
	for node, timestamp := range members.Members {

		duration := time.Since(timestamp)

		if duration.Seconds() > 30 {
			log.Println("Tried to add an expired node", node)
			continue
		}

		self.Members[node] = timestamp
	}
}

func (self *Membership) Random() (string, time.Time) {
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
