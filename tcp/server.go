package tcp

import (
	"net"

	"github.com/Sirupsen/logrus"
	msg "github.com/UniversityRadioYork/bifrost-go/message"
)

type Server struct {
	addr    string
	clients map[*Client]bool
	log     *logrus.Logger
}

// Broadcast sends a message to all connected clients.
func (s *Server) Broadcast(message *msg.Message) {
	for client := range s.clients {
		client.Send(message)
	}
}

// Listen starts listening on the previously specified address for new Client
// connections. New connections are sent down newConn channel.
func (s *Server) Listen(newConn chan<- net.Conn) {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Fatal(err)
	}
	s.log.Info("Listening on ", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log.Warn(err)
		}
		newConn <- conn
	}
}

// RemoveClient closes Client c's connection and removes it from the map.
func (s *Server) RemoveClient(c *Client) error {
	// TODO: Do this here?
	err := c.Close()
	if err != nil {
		// Err, help?
		return err
	}
	delete(s.clients, c)
	return nil
}

// AddClient adds Client c to the map of clients.
func (s *Server) AddClient(c *Client) {
	s.clients[c] = true
}

// NewServer creates a new Server instance with the host & port to listen on,
// and a logger instance.
func NewServer(hostport string, logger *logrus.Logger) *Server {
	s := &Server{
		addr:    hostport,
		clients: make(map[*Client]bool),
		log:     logger,
	}
	return s
}
