package main

import (
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

// Maintains communications with the connector and connected clients.
// Also does any processing needed with the commands.
type hub struct {
	// All current connections, and their outbound channel.
	clients map[net.Conn]chan<- baps3.Message

	// For communication with the connector.
	cReqCh chan<- baps3.Message
	cResCh <-chan baps3.Message

	// Where new requests from clients come through.
	reqCh chan baps3.Message

	// Handlers for adding/removing connections.
	addCh chan *Client
	rmCh  chan *Client
}

var h = hub{
	clients: make(map[net.Conn]chan<- baps3.Message),

	reqCh: make(chan baps3.Message),

	addCh: make(chan *Client),
	rmCh:  make(chan *Client),
}

// Handles a new client connection.
// conn is the new connection object.
func (h *hub) handleNewConnection(conn net.Conn) {
	defer conn.Close()
	client := &Client{
		conn:  conn,
		resCh: make(chan baps3.Message),
		tok:   baps3.NewTokeniser(),
	}

	// Register user
	h.addCh <- client

	go client.Read(h.reqCh, h.rmCh)
	client.Write(client.resCh, h.rmCh)
}

func makeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd-" + LD_VERSION)
}

func makeFeaturesMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsFeatures).AddArg("lol")
}

// Handles a request from a client.
// Falls through to the connector cReqCh if command is "not understood".
func (h *hub) processRequest(req baps3.Message) {
	// TODO: Do something else
	log.Println("New request:", req.String())
	h.cReqCh <- req
}

// Broadcasts a response (res) to all connected clients.
func (h *hub) processResponse(res baps3.Message) {
	// TODO: Do something else
	log.Println("New response:", res.String())
	for _, ch := range h.clients {
		ch <- res
	}
}

// Listens for new connections on addr:port and spins up the relevant goroutines.
func (h *hub) runListener(addr string, port string) {
	netListener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		log.Println("Listening error:", err.Error())
		return
	}

	// Get new connections
	go func() {
		for {
			conn, err := netListener.Accept()
			if err != nil {
				log.Println("Error accepting connection:", err.Error())
				continue
			}

			go h.handleNewConnection(conn)
		}
	}()

	for {
		select {
		case msg := <-h.cResCh:
			h.processResponse(msg)
		case msg := <-h.reqCh:
			h.processRequest(msg)
		case client := <-h.addCh:
			h.clients[client.conn] = client.resCh
			client.resCh <- *makeWelcomeMsg()
			client.resCh <- *makeFeaturesMsg()
			log.Println("New connection from", client.conn.RemoteAddr())
		case client := <-h.rmCh:
			close(client.resCh)
			delete(h.clients, client.conn)
			log.Println("Closed connection from", client.conn.RemoteAddr())
		}
	}
}

// Sets up the connector channels for the hub object.
func (h *hub) setConnector(cReqCh chan<- baps3.Message, cResCh <-chan baps3.Message) {
	h.cReqCh = cReqCh
	h.cResCh = cResCh
}
