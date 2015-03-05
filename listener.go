package main

import (
	"bufio"
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type Client struct {
	conn  net.Conn
	resCh chan baps3.Message
	tok   *baps3.Tokeniser
}

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

func (c *Client) Read(reqCh chan<- baps3.Message, rmCh chan<- *Client) {
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
			reqCh <- *msg
		}
	}
}

func (c *Client) Write(ch <-chan baps3.Message, rmCh chan<- *Client) {
	for {
		msg, more := <-ch
		// Channel's been closed
		if !more {
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

func MakeWelcomeMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsOhai).AddArg("listd").AddArg("0.0")
}

func MakeFeaturesMsg() *baps3.Message {
	return baps3.NewMessage(baps3.RsFeatures).AddArg("lol")
}

func ProcessRequest(connectorReqCh chan<- baps3.Message, req baps3.Message) {
	// TODO: Do something else
	log.Println("New request:", req.String())
	connectorReqCh <- req
}

func ProcessResponse(clients *map[net.Conn]chan<- baps3.Message, res baps3.Message) {
	// TODO: Do something else
	log.Println("New response:", res.String())
	for _, ch := range *clients {
		ch <- res
	}
}

func handleChannels(reqCh <-chan baps3.Message, cReqCh chan<- baps3.Message, cResCh <-chan baps3.Message, addCh <-chan *Client, rmCh <-chan *Client) {

	clients := make(map[net.Conn]chan<- baps3.Message)

	for {
		select {
		case msg := <-cResCh:
			ProcessResponse(&clients, msg)
		case msg := <-reqCh:
			ProcessRequest(cReqCh, msg)
		case client := <-addCh:
			clients[client.conn] = client.resCh
			client.resCh <- *MakeWelcomeMsg()
			client.resCh <- *MakeFeaturesMsg()
			log.Println("New connection from", client.conn.RemoteAddr())
		case client := <-rmCh:
			close(client.resCh)
			delete(clients, client.conn)
			log.Println("Closed connection from", client.conn.RemoteAddr())
		}
	}
}

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
