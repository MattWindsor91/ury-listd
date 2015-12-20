package main

import (
	"net"
	"strconv"
	"strings"

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

func OhaiMsg() msg.Message {
	return msg.Message{msg.RsOhai, ProgVersion, "bifrost α"}
}

func (ctx *Context) onClientConnect(c *tcp.Client) {
	ctx.log.Info("New client connection ", c.RemoteAddr())
	c.Send(OhaiMsg())
	// Logic goes here
	// TODO: Send tree?
}

func (ctx *Context) onClientDisconnect(c *tcp.Client, err error) {
	if rmErr := ctx.s.RemoveClient(c); rmErr != nil {
		// Err, help?
		ctx.log.Warn("Error closing connection: ", rmErr)
	}
	ctx.log.Warn("Connection lost from ", c.RemoteAddr(), " because ", err)
	// Logic goes here
}

func splitPath(path string) []string {
	f := func(c rune) bool {
		return c == '/'
	}
	return strings.FieldsFunc(path, f)
}

func joinPath(path []string) string {
	return "/" + strings.Join(path, "/")
}

func msgLenCheck(cmd msg.Message) bool {
	l := len(cmd)
	if l < 1 {
		return false
	}
	switch cmd[0] {
	case msg.RqRead:
	case msg.RqDelete:
		return l == 3
	case msg.RqWrite:
		return l == 4
	case msg.RsRes:
		return l == 5
	case msg.RsUpdate:
		return l == 4
	case msg.RsAck:
		// Pfft, who knows.
		return l > 3
	}
	return false
}

func (ctx *Context) onNewRequest(c *tcp.Client, message msg.Message) msg.Message {
	ctx.log.Debug("Request: ", message.String())

	if !msgLenCheck(message) {
		return msg.Ack(msg.AckWhat, "Bad command or file name", message)
	}

	tag := message[1]
	path := splitPath(message[2])

	if len(path) != 3 || path[0] != "playlist" {
		// If it's not to do with playlists, we don't care about it and can send it on
		ctx.outgoing.Send(message)
		return nil
	}

	idx, err := strconv.Atoi(path[1])
	if err != nil {
		return msg.Ack(msg.AckFail, "Index must be numeric", message)
	}

	//	hash := path[2]
	switch message[0] {
	case msg.RqRead:
		item, err := ctx.pl.Get(idx)
		if err != nil {
			return msg.Ack(msg.AckFail, err.Error(), message)
		}
		c.Send(msg.Res(tag, joinPath(path), "string", item.Data))
		return msg.Ack(msg.AckOk, "success", message)
	case msg.RqWrite:
	case msg.RqDelete:
	default:
		return msg.Ack(msg.AckWhat, "Bad command or file name", message)
	}

	// Not something we recognise, pass it on for something else to fail.
	ctx.outgoing.Send(message)
	return nil
}

func isOurTag(tag string) bool {
	return strings.HasPrefix(tag, "listd")
}

func (ctx *Context) onNewResponse(message msg.Message) {
	ctx.log.Debug("Response: ", message.String())

	// Invalid or not our response, pass it on.
	if len(message) < 2 {
		return
	} else if (message[0] == msg.RsAck || message[0] == msg.RsRes) && !isOurTag(message[1]) {
		ctx.s.Broadcast(message)
	}
	// logic

	// if update, pass on
	if message[0] == msg.RsUpdate {
		ctx.s.Broadcast(message)
	}
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
			ackMsg := ctx.onNewRequest(request.C, request.M)
			if ackMsg != nil {
				request.C.Send(ackMsg)
			}

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
