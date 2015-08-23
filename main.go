package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/UniversityRadioYork/ury-listd-go/tcpserver"
	"github.com/docopt/docopt-go"
)

// Version string, provided by linker flags
var LDVersion string

func parseArgs() (args map[string]interface{}, err error) {
	usage := `ury-listd-go.

Usage:
  ury-listd-go [-p <port>] [-a <address>] [-P <port>] [-A <address>]
  ury-listd-go -h
  ury-listd-go -v

Options:
  -p --port=<port>              The port ury-listd-go listens on [default: 1351].
  -a --addr=<address>           The host ury-listd-go listens on [default: 127.0.0.1].
  -P --playoutport=<port>       The playout system's listening port [default: 1350].
  -A --playoutaddr=<address>    The playout system's listening address [default: 127.0.0.1].
  -h --help                     Show this screen.
  -v --version                  Show version.`

	return docopt.Parse(usage, nil, true, "ury-listd-go "+LDVersion, false)
}

// TODO Rename?
type context struct {
	log    *logrus.Logger
	server *tcpserver.Server
}

func (ctxt *context) onClientConnect(c *tcpserver.Client) {
	ctxt.log.Info("New client: ", c.RemoteAddr())
}

func (ctxt *context) onClientDisconnect(c *tcpserver.Client, err error) {
	ctxt.log.Warn("Client gone: ", c.RemoteAddr(), " because ", err)
}

func (ctxt *context) onNewMessage(c *tcpserver.Client, message string) {
	ctxt.log.Info("Msg: ", message)
	ctxt.server.Broadcast(message + "\n")
}

func main() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	// Get arguments (or their defaults)
	args, err := parseArgs()
	if err != nil {
		logger.Fatal("Error parsing args: " + err.Error())
	}

	hostport := args["--addr"].(string) + ":" + args["--port"].(string)
	s := tcpserver.New(hostport)

	// TODO: Make connection to playout system

	c := &context{
		log:    logger,
		server: s,
	}

	s.SetClientConnectFunc(c.onClientConnect)
	s.SetClientDisconnectFunc(c.onClientDisconnect)
	s.SetNewMessageFunc(c.onNewMessage)

	logger.Info("Listening on ", hostport)
	if err := s.Listen(); err != nil {
		logger.Error(err)
	}

}
