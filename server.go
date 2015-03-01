package main

import (
	"log"
	"net"

	baps3 "github.com/UniversityRadioYork/baps3-go"
)

type Server struct {
	listener   net.Listener
	serverCh   chan baps3.Message
	clientComm chan []byte
	tok        *baps3.Tokeniser
	clients    []*Client
}

func MakeServer(addr string, port string, serverCh chan baps3.Message) (*Server, error) {
	listener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener:   listener,
		tok:        baps3.NewTokeniser(),
		clientComm: make(chan []byte),
		serverCh:   serverCh,
	}
	return server, nil
}

func (s *Server) Broadcast(msg baps3.Message) {
	// Probably should have \n at the end
	for _, c := range s.clients {
		data, err := msg.Pack()
		if err != nil {
			continue // TODO Handle
		}
		c.Outgoing <- data
	}
}

func (s *Server) ReceiveLoop() {
	for data := range s.clientComm {
		_, _, err := s.tok.Tokenise(data)
		if err != nil {
			continue // TODO: Do something?
		}
	}
}

func (s *Server) run() {
	go s.ReceiveLoop()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		client := MakeClient(conn, s.clientComm)
		s.clients = append(s.clients, client)
	}
}
