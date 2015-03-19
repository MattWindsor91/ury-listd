package main

import (
	"bufio"
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

// Wrapper structure for a client connection. The actual connection is stored in conn,
// resCh is a channel that responses get sent down and tok is the tokeniser for
// converting newly received data into baps3.Messages.
type Client struct {
	conn  net.Conn
	resCh chan baps3.Message
	tok   *baps3.Tokeniser
}

// Reads data from a client connection. All received request messages get sent down reqCh.
// Bails if reading bytes causes an error, which gets the connection unregistered and disconnected.
func (c *Client) Read(reqCh chan<- clientAndMessage, rmCh chan<- *Client) {
	reader := bufio.NewReader(c.conn)
	for {
		// Get new request
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println("Error reading from", c.conn.RemoteAddr(), ":", err.Error())
			rmCh <- c
			return
		}
		lines, _, err := c.tok.Tokenise(line)
		if err != nil {
			log.Println(err)
			continue // TODO: Do something?
		}
		for _, line := range lines {
			msg, err := baps3.LineToMessage(line)
			if err != nil {
				log.Println(err)
				continue // TODO: Do something?
			}
			reqCh <- clientAndMessage{c, *msg}
		}
	}
}

// Writes new responses to the client connection.
// New responses are got from resCh. Errors in writing the data
// will cause the connection to be disconnected, via rmCh.
func (c *Client) Write(resCh <-chan baps3.Message, rmCh chan<- *Client) {
	for {
		msg, ok := <-resCh
		// Channel's been closed
		if !ok {
			return
		}
		data, err := msg.Pack()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		_, err = c.conn.Write(data)
		if err != nil {
			log.Println("Error writing from", c.conn.RemoteAddr(), ":", err.Error())
			rmCh <- c
			return
		}
	}
}
