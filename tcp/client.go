package tcp

import (
	"bufio"
	"log"
	"net"
)

// Client is a connection to our TCP server.
type Client struct {
	net.Conn
	rw    *bufio.ReadWriter
	msgCh chan<- ClientMessage
	rmCh  chan<- ClientError
}

// ClientError is a tuple type for sending client and an error down a channel.
type ClientError struct {
	C *Client
	E error
}

// ClientMessage is a tuple type for sending client and message down a channel.
type ClientMessage struct {
	C *Client
	M []byte
}

// Listen reads new lines from the socket connection. A new message gets sent
// down the msgCh channel. If there is an error in reading (usually indicating
// the connection died), the function sends the error and itself down the rmCh
// and returns.
func (c *Client) Listen() {
	for {
		message, err := c.rw.ReadBytes('\n')
		if err != nil {
			c.rmCh <- ClientError{c, err} // Remove self
			return
		}
		c.msgCh <- ClientMessage{c, message}
	}
}

// Send asyncronously writes a message string to the client instance.
func (c *Client) Send(message []byte) {
	if _, err := c.rw.Write(message); err != nil {
		// If error, reasonable to assume client has been removed by listen
		// gorountine. No need to do anything else.
		log.Println(err)
	}
	if err := c.rw.Flush(); err != nil {
		// Same as above
		log.Println(err)
	}
}

// NewClient creates a new client instance from an existing connection conn,
// a request channel to send new messages down and an removal channel to send
// errors down when there's an error reading.
func NewClient(conn net.Conn, requestCh chan<- ClientMessage, rmCh chan<- ClientError) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &Client{
		Conn:  conn,
		rw:    bufio.NewReadWriter(reader, writer),
		msgCh: requestCh,
		rmCh:  rmCh,
	}
	return client
}
