package main

import (
	"net"

	"github.com/Sirupsen/logrus"
	msg "github.com/UniversityRadioYork/bifrost-go/message"
	"github.com/UniversityRadioYork/ury-listd/tcp"
)

// Context is the main listd instance. Manages TCP connections (including the
// connector) and controls the Playlist instance.
type Context struct {
	log *logrus.Logger
	pl  *Playlist
	s   *tcp.Server

	// TCP Stuffs
	outgoingAddr string
	outgoing     *tcp.Client
}

func (ctx *Context) onClientConnect(c *tcp.Client) {
	ctx.log.Info("New client connection ", c.RemoteAddr())
	// Logic goes here
}

func (ctx *Context) onClientDisconnect(c *tcp.Client, err error) {
	if rmErr := ctx.s.RemoveClient(c); rmErr != nil {
		// Err, help?
		ctx.log.Warn("Error closing connection: ", rmErr)
	}
	ctx.log.Warn("Connection lost from ", c.RemoteAddr(), " because ", err)
	// Logic goes here
}

func (ctx *Context) onNewRequest(c *tcp.Client, message *msg.Message) {
	ctx.log.Debug("Request: ", message.String())
	// Logic goes here
	ctx.outgoing.Send(message)
}

func (ctx *Context) onNewResponse(message *msg.Message) {
	ctx.log.Debug("Response: ", message.String())
	// Logic goes here
	ctx.s.Broadcast(message)
}

// Run the main loop (handling channels).
func (ctx *Context) Run() {
	newCh := make(chan net.Conn)
	go ctx.s.Listen(newCh)

	// Make connection
	responseCh := make(chan tcp.ClientMessage)
	outgoingRmCh := make(chan tcp.ClientError)
	conn, err := net.Dial("tcp", ctx.outgoingAddr)
	if err != nil {
		ctx.log.Fatal(err)
	}
	ctx.outgoing = tcp.NewClient(conn, responseCh, outgoingRmCh)
	go ctx.outgoing.Listen()

	// Main loop
	requestCh := make(chan tcp.ClientMessage)
	rmCh := make(chan tcp.ClientError)
	for {
		select {
		// Client stuff
		case conn := <-newCh:
			client := tcp.NewClient(conn, requestCh, rmCh)
			ctx.s.AddClient(client)
			go client.Listen()
			ctx.onClientConnect(client)
		case clienterr := <-rmCh:
			ctx.onClientDisconnect(clienterr.C, clienterr.E)
		case request := <-requestCh:
			ctx.onNewRequest(request.C, request.M)

		// Outgoing stuff
		case response := <-responseCh:
			// Don't care about the client, we already know about that
			ctx.onNewResponse(response.M)
		case outgoingerr := <-outgoingRmCh:
			// Connection died :(
			// TODO: Handle it
			ctx.log.Fatal(outgoingerr.E)
			// TODO: Stop the world.
		}
	}
}

// NewContext creates a new context that will listen for clients on hostport,
// make an outgoing connection on outgoing and make use of logger for logging.
func NewContext(hostport, outgoing string, logger *logrus.Logger) *Context {
	c := &Context{
		outgoingAddr: outgoing,
		log:          logger,
		s:            tcp.NewServer(hostport, logger),
	}
	return c
}
