package main

import (
	"bufio"
	"net"
)

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	Incoming chan string
	Outgoing chan string
}

func MakeClient(conn net.Conn) *Client {
	client := &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		Incoming: make(chan string),
		Outgoing: make(chan string),
	}

	go client.Read()
	go client.Write()

	client.Outgoing <- "Hello!\n" // XXX Actual welcome message

	return client
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Read() {
	for {
		line, _ := c.reader.ReadString('\n')
		c.Incoming <- line
	}
}

func (c *Client) Write() {
	for data := range c.Outgoing {
		c.writer.WriteString(data)
		c.writer.Flush()
	}
}

type Server struct {
	listener net.Listener
	serverCh chan string
	clients  []*Client
}

func MakeServer(addr string, port string, serverCh chan string) (*Server, error) {
	listener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		serverCh: serverCh,
	}
	return server, nil
}

func (s *Server) Write(msg string) {
	// Probably should have \n at the end
	for _, c := range s.clients {
		c.Outgoing <- msg
	}
}

func (s *Server) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		client := MakeClient(conn)
		s.clients = append(s.clients, client)
	}
}
