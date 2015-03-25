package main

import (
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

// Maintains communications with the downstream service and connected clients.
// Also does any processing needed with the commands.
type hub struct {
	// All current clients.
	clients map[*Client]bool

	// Dump state from the downstream service (playd)
	downstreamVersion  string
	downstreamFeatures baps3.FeatureSet

	// For communication with the downstream service.
	cReqCh chan<- baps3.Message
	cResCh <-chan baps3.Message

	// Where new requests from clients come through.
	reqCh chan baps3.Message

	// Handlers for adding/removing connections.
	addCh chan *Client
	rmCh  chan *Client
	Quit  chan bool
}

var h = hub{
	clients: make(map[*Client]bool),

	downstreamFeatures: make(baps3.FeatureSet),

	reqCh: make(chan baps3.Message),

	addCh: make(chan *Client),
	rmCh:  make(chan *Client),
	Quit:  make(chan bool),
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

// Appends the downstream service's version (from the OHAI) to the listd version.
func makeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd " + LD_VERSION + "/" + h.downstreamVersion)
}

// Crafts the features message by adding listd's features to the downstream service's and removing
// features listd intercepts.
func makeFeaturesMsg() (msg *baps3.Message) {
	features := h.downstreamFeatures
	features.DelFeature(baps3.FtFileLoad) // 'Mask' the features listd intercepts
	features.AddFeature(baps3.FtPlaylist)
	features.AddFeature(baps3.FtPlaylistTextItems)
	msg = features.ToMessage()
	return
}

// Handles a request from a client.
// Falls through to the connector cReqCh if command is "not understood".
func (h *hub) processRequest(req baps3.Message) {
	// TODO: Do something else
	log.Println("New request:", req.String())
	h.cReqCh <- req
}

// Processes a response from the downstream service.
func (h *hub) processResponse(res baps3.Message) {
	// TODO: Do something else
	log.Println("New response:", res.String())
	switch res.Word() {
	case baps3.RsOhai:
		h.downstreamVersion, _ = res.Arg(0)
	case baps3.RsFeatures:
		fs, err := baps3.FeatureSetFromMsg(&res)
		if err != nil {
			log.Fatal("Error reading features: " + err.Error())
		}
		h.downstreamFeatures = fs
	default:
		h.broadcast(res)
	}
}

// Send a response message to all clients.
func (h *hub) broadcast(res baps3.Message) {
	for c, _ := range h.clients {
		c.resCh <- res
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
			h.clients[client] = true
			client.resCh <- *makeWelcomeMsg()
			client.resCh <- *makeFeaturesMsg()
			log.Println("New connection from", client.conn.RemoteAddr())
		case client := <-h.rmCh:
			close(client.resCh)
			delete(h.clients, client)
			log.Println("Closed connection from", client.conn.RemoteAddr())
		case <-h.Quit:
			log.Println("Closing all connections")
			for c, _ := range h.clients {
				close(c.resCh)
				delete(h.clients, c)
			}
			//			h.Quit <- true
		}
	}
}

// Sets up the connector channels for the hub object.
func (h *hub) setConnector(cReqCh chan<- baps3.Message, cResCh <-chan baps3.Message) {
	h.cReqCh = cReqCh
	h.cResCh = cResCh
}
