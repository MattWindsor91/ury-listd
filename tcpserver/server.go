package tcpserver

import (
	"log"
	"net"
)

// Server is the main TCP Server instance. Manages clients and provides hooks
// that are called whenever a client connects/disconnects, or when a new
// message arrives.
type Server struct {
	addr    string
	clients map[*Client]bool

	onClientConnectCB    func(c *Client)
	onClientDisconnectCB func(c *Client, err error)
	onNewMessageCB       func(c *Client, message []byte)
}

func (s *Server) onClientConnect(c *Client) {
	s.onClientConnectCB(c)
}

func (s *Server) onClientDisconnect(c *Client, err error) {
	clErr := c.Close()
	if clErr != nil {
		// Err, help?
		log.Println("Error closing connection:", clErr)
	}
	delete(s.clients, c)
	s.onClientDisconnectCB(c, err)
}

func (s *Server) onNewMessage(c *Client, message []byte) {
	s.onNewMessageCB(c, message)
}

// SetClientConnectFunc sets the user's endpoint func for handling a new
// client.
func (s *Server) SetClientConnectFunc(cb func(c *Client)) {
	s.onClientConnectCB = cb
}

// SetClientDisconnectFunc sets the user's endpoint func for handling a client
// disconnect.
func (s *Server) SetClientDisconnectFunc(cb func(c *Client, err error)) {
	s.onClientDisconnectCB = cb
}

// SetNewMessageFunc sets the user's endpoint func for handling a new message.
func (s *Server) SetNewMessageFunc(cb func(c *Client, message []byte)) {
	s.onNewMessageCB = cb
}

// Broadcast sends a message to all connected clients.
func (s *Server) Broadcast(message []byte) {
	for client := range s.clients {
		client.Send(message)
	}
}

func (s *Server) newClient(conn net.Conn, rmCh chan<- clientError, msgCh chan<- clientMessage) {
	client := &Client{
		Conn: conn,
	}
	s.clients[client] = true
	s.onClientConnect(client)
	go client.listen(rmCh, msgCh)
}

func (s *Server) listenNewConnections(newConn chan<- net.Conn) {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		newConn <- conn
	}
}

// Listen for new connections and handle channels
func (s *Server) Listen() {
	newCh := make(chan net.Conn)
	rmCh := make(chan clientError)
	msgCh := make(chan clientMessage)

	go s.listenNewConnections(newCh)

	for {
		select {
		case conn := <-newCh:
			s.newClient(conn, rmCh, msgCh)
		case clienterr := <-rmCh:
			s.onClientDisconnect(clienterr.c, clienterr.e)
		case clientmsg := <-msgCh:
			s.onNewMessage(clientmsg.c, clientmsg.m)
			// TODO: Stop the world.
		}
	}
}

// New creates a new server instance that listens on hostport.
func New(hostport string) *Server {
	s := &Server{
		addr:                 hostport,
		clients:              make(map[*Client]bool),
		onClientConnectCB:    func(c *Client) {},
		onClientDisconnectCB: func(c *Client, err error) {},
		onNewMessageCB:       func(c *Client, message []byte) {},
	}
	return s
}
