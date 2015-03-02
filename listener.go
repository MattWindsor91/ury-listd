package main

import (
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type Listener struct {
	listener   net.Listener
	requestCh  chan baps3.Message
	responseCh chan []byte
	tok        *baps3.Tokeniser
	clients    []*Client
}

func MakeListener(addr string, port string, reqCh chan baps3.Message) (*Listener, error) {
	netListener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}
	listener := &Listener{
		listener:   netListener,
		tok:        baps3.NewTokeniser(),
		responseCh: make(chan []byte),
		requestCh:  reqCh,
	}
	return listener, nil
}

func (l *Listener) Broadcast(msg baps3.Message) {
	// Probably should have \n at the end
	for _, c := range l.clients {
		data, err := msg.Pack()
		if err != nil {
			continue // TODO Handle
		}
		c.Outgoing <- data
	}
}

func (l *Listener) ReceiveLoop() {
	for data := range l.responseCh {
		lines, _, err := l.tok.Tokenise(data)
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
			// TODO: Something with particular types of msgs
			l.requestCh <- *msg
		}
	}
}

func (s *Listener) ProcessCommand(msg baps3.Message) {
	s.Broadcast(msg)
}

func (l *Listener) run() {
	go l.ReceiveLoop()
	for {
		conn, err := l.listener.Accept()
		log.Println("Opening connection from " + conn.RemoteAddr().String())
		if err != nil {
			log.Println("Error accepting connection: " + err.Error())
			continue
		}
		client := MakeClient(conn, l.responseCh)
		l.clients = append(l.clients, client)
	}
}
