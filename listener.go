package main

import (
	"log"
	"net"
	"strconv"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type clientAndMessage struct {
	c   *Client
	msg baps3.Message
}

// Maintains communications with the downstream service and connected clients.
// Also does any processing needed with the commands.
type hub struct {
	// All current clients.
	clients map[*Client]bool

	// Dump state from the downstream service (playd)
	downstreamVersion  string
	downstreamFeatures baps3.FeatureSet

	// Playlist instance
	pl *Playlist

	// For communication with the downstream service.
	cReqCh chan<- baps3.Message
	cResCh <-chan baps3.Message

	// Where new requests from clients come through.
	reqCh chan clientAndMessage

	// Handlers for adding/removing connections.
	addCh chan *Client
	rmCh  chan *Client
	Quit  chan bool
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
func (h *hub) makeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd " + LD_VERSION + "/" + h.downstreamVersion)
}

// Crafts the features message by adding listd's features to the downstream service's and removing
// features listd intercepts.
func (h *hub) makeFeaturesMsg() (msg *baps3.Message) {
	features := h.downstreamFeatures
	features.DelFeature(baps3.FtFileLoad) // 'Mask' the features listd intercepts
	features.AddFeature(baps3.FtPlaylist)
	features.AddFeature(baps3.FtPlaylistTextItems)
	msg = features.ToMessage()
	return
}

func sendInvalidCmd(c *Client, errRes baps3.Message, oldCmd baps3.Message) {
	for _, w := range oldCmd.AsSlice() {
		errRes.AddArg(w)
	}
	c.resCh <- errRes
}

func processReqDequeue(pl *Playlist, req baps3.Message) *baps3.Message {
	args := req.AsSlice()[1:]
	if len(args) != 2 {
		return baps3.NewMessage(baps3.RsWhat).AddArg("Bad command")
	}
	iStr, hash := args[0], args[1]

	i, err := strconv.Atoi(iStr)
	if err != nil {
		return baps3.NewMessage(baps3.RsWhat).AddArg("Bad index")
	}

	rmIdx, rmHash, err := pl.Dequeue(i, hash)
	if err != nil {
		return baps3.NewMessage(baps3.RsFail).AddArg(err.Error())
	}
	return baps3.NewMessage(baps3.RsDequeue).AddArg(strconv.Itoa(rmIdx)).AddArg(rmHash)
}

func processReqEnqueue(pl *Playlist, req baps3.Message) *baps3.Message {
	args := req.AsSlice()[1:]
	if len(args) != 4 {
		return baps3.NewMessage(baps3.RsWhat).AddArg("Bad command")
	}
	iStr, hash, itemType, data := args[0], args[1], args[2], args[3]

	i, err := strconv.Atoi(iStr)
	if err != nil {
		return baps3.NewMessage(baps3.RsWhat).AddArg("Bad index")
	}

	if itemType != "file" && itemType != "text" {
		return baps3.NewMessage(baps3.RsWhat).AddArg("Bad item type")
	}

	item := &PlaylistItem{Data: data, Hash: hash, IsFile: itemType == "file"}
	newIdx, err := pl.Enqueue(i, item)
	if err != nil {
		return baps3.NewMessage(baps3.RsFail).AddArg(err.Error())
	}
	return baps3.NewMessage(baps3.RsEnqueue).AddArg(strconv.Itoa(newIdx)).AddArg(item.Hash).AddArg(itemType).AddArg(item.Data)

}

var REQ_MAP = map[baps3.MessageWord]func(*Playlist, baps3.Message) *baps3.Message{
	baps3.RqDequeue: processReqDequeue,
	baps3.RqEnqueue: processReqEnqueue,
}

// Handles a request from a client.
// Falls through to the connector cReqCh if command is "not understood".
func (h *hub) processRequest(c *Client, req baps3.Message) {
	log.Println("New request:", req.String())
	if reqFunc, ok := REQ_MAP[req.Word()]; ok {
		// TODO: Add a "is fail word" func to baps3-go?
		if resp := reqFunc(h.pl, req); resp.Word() == baps3.RsFail || resp.Word() == baps3.RsWhat {
			// failures only go to sender
			sendInvalidCmd(c, *resp, req)
		} else {
			h.broadcast(*resp)
		}
	} else {
		h.cReqCh <- req
	}
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
		case data := <-h.reqCh:
			h.processRequest(data.c, data.msg)
		case client := <-h.addCh:
			h.clients[client] = true
			client.resCh <- *h.makeWelcomeMsg()
			client.resCh <- *h.makeFeaturesMsg()
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
