package main

import (
	"log"
	"sync"
)

type hub struct {
	mu      sync.Mutex
	clients []*client
}

func (b *hub) sub(c *client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients = append(b.clients, c)
}

func (b *hub) usub(c *client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.clients[:0]
	for _, x := range b.clients {
		if x != c {
			n = append(n, x)
		}
	}
	b.clients = n
}

func (b *hub) pub(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.clients {
		// send but do not block for it
		select {
		case c.c <- v:
		default:
			log.Printf("failed to broadcast %v to %v as the receiving channel is busy\n", v, c)
		}
	}
}

func (b *hub) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.clients {
		c.close()
	}
}
