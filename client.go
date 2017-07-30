package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	conn   *websocket.Conn
	c      chan interface{}
	mu     sync.Mutex
	closed bool
}

func (c *client) serve() {
	go func() {
		for {
			if _, _, err := c.conn.NextReader(); err != nil {
				b.usub(c)
				c.close()
				break
			}
		}
	}()

	for v := range c.c {
		if err := c.conn.WriteJSON(v); err != nil {
			log.Println(err)
		}
	}
}

func (c *client) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	close(c.c)

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close a client: %v", err)
	}
	return nil
}
