package main

import (
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

// Handles a new client connection.
// reqCh is the channel that receives the new requests from the connection and
// addCh & rmCh are for (un)registering the channel with the main client list.
func handleNewConnection(conn net.Conn, reqCh chan<- baps3.Message, addCh chan<- *Client, rmCh chan<- *Client) {
	defer conn.Close()
	client := &Client{
		conn:  conn,
		resCh: make(chan baps3.Message),
		tok:   baps3.NewTokeniser(),
	}

	// Register user
	addCh <- client

	go client.Read(reqCh, rmCh)
	client.Write(client.resCh, rmCh)
}

func makeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd").AddArg("0.0")
}

func makeFeaturesMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsFeatures).AddArg("lol")
}

// Handles a request from a client.
// Falls through to the connector cReqCh if command is "not understood".
func processRequest(cReqCh chan<- baps3.Message, req baps3.Message) {
	// TODO: Do something else
	log.Println("New request:", req.String())
	cReqCh <- req
}

// Broadcasts a response (res) to all connected clients.
func processResponse(clients *map[net.Conn]chan<- baps3.Message, res baps3.Message) {
	// TODO: Do something else
	log.Println("New response:", res.String())
	for _, ch := range *clients {
		ch <- res
	}
}

// Main handler for client connection channels.
// reqCh is the channel that new requests from clients come through.
// cReqCh & cResCh are channels from the connector - cReqCh getting the "unused" requests
// and cResCh being polled for responses from the connector.
// addCh & rmCh are handlers for adding/removing connections.
func handleChannels(reqCh <-chan baps3.Message, cReqCh chan<- baps3.Message, cResCh <-chan baps3.Message, addCh <-chan *Client, rmCh <-chan *Client) {

	clients := make(map[net.Conn]chan<- baps3.Message)

	for {
		select {
		case msg := <-cResCh:
			processResponse(&clients, msg)
		case msg := <-reqCh:
			processRequest(cReqCh, msg)
		case client := <-addCh:
			clients[client.conn] = client.resCh
			client.resCh <- *makeWelcomeMsg()
			client.resCh <- *makeFeaturesMsg()
			log.Println("New connection from", client.conn.RemoteAddr())
		case client := <-rmCh:
			close(client.resCh)
			delete(clients, client.conn)
			log.Println("Closed connection from", client.conn.RemoteAddr())
		}
	}
}

// Listens for new connections on addr:port and spins up the relevant goroutines.
// cReqCh & cResCh are from the connector, requests get pushed down and responses get pulled, respectively.
func runListener(addr string, port string, cReqCh chan<- baps3.Message, cResCh <-chan baps3.Message) {
	netListener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		log.Println("Listening error:", err.Error())
		return
	}

	reqCh := make(chan baps3.Message)
	addCh := make(chan *Client)
	remCh := make(chan *Client)

	go handleChannels(reqCh, cReqCh, cResCh, addCh, remCh)

	// Get new connections
	for {
		conn, err := netListener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleNewConnection(conn, reqCh, addCh, remCh)
	}
}
