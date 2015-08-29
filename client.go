package main

import (
	"bufio"
	"log"
	"net"
)

// Client is a connection to our TCP server.
type Client struct {
	net.Conn
}

// Tuple type for sending client and an error down a channel.
type clientError struct {
	c *Client
	e error
}

// Tuple type for sending client and a message down a channel.
type clientMessage struct {
	c *Client
	m []byte
}

func (c *Client) listen(rmCh chan<- clientError, msgCh chan<- clientMessage) {
	reader := bufio.NewReader(c)
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			rmCh <- clientError{c, err} // Remove self
			return
		}
		msgCh <- clientMessage{c, message}
	}
}

// Send asyncronously writes a message string to the client instance.
func (c *Client) Send(message []byte) {
	// We don't care about any return value (errors handled by listen thread)
	// so run this in a goroutine to stop slow things (e.g. networks) slowing
	// the whole program down.
	go func() {
		_, err := c.Write(message)
		if err != nil {
			// If error, reasonable to assume client has been removed by listen
			// gorountine. No need to do anything else.
			log.Println(err)
		}
	}()
}
