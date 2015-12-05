package tcp

import (
	"log"
	"net"

	msg "github.com/UniversityRadioYork/bifrost-go/message"
	"github.com/UniversityRadioYork/bifrost-go/tokeniser"
)

// Client is a connection to our TCP server.
type Client struct {
	net.Conn
	tok   *tokeniser.Tokeniser
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
	M msg.Message
}

// Listen reads new lines from the socket connection. A new message gets sent
// down the msgCh channel. If there is an error in reading (usually indicating
// the connection died), the function sends the error and itself down the rmCh
// and returns.
func (c *Client) Listen() {
	for {
		line, err := c.tok.Tokenise()
		if err != nil {
			c.rmCh <- ClientError{c, err} // Remove self
			return
		}

		message := msg.Message(line)
		c.msgCh <- ClientMessage{c, message}
	}
}

// Send asyncronously writes a message string to the client instance.
func (c *Client) Send(message msg.Message) {
	data := message.Pack()

	if _, err := c.Write(data); err != nil {
		// If error, reasonable to assume client has been removed by listen
		// gorountine. No need to do anything else.
		log.Println(err)
	}
}

// NewClient creates a new client instance from an existing connection conn,
// a request channel to send new messages down and an removal channel to send
// errors down when there's an error reading.
func NewClient(conn net.Conn, requestCh chan<- ClientMessage, rmCh chan<- ClientError) *Client {
	client := &Client{
		Conn:  conn,
		tok:   tokeniser.New(conn),
		msgCh: requestCh,
		rmCh:  rmCh,
	}
	return client
}
