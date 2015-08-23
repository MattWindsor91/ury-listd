package main

import (
	"bytes"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/UniversityRadioYork/ury-listd-go/tcpserver"
	"github.com/docopt/docopt-go"
)

// Version string, provided by linker flags
var LDVersion string

// TODO Rename?
type context struct {
	log      *logrus.Logger
	server   *tcpserver.Server
	playout  *Connection
	playlist *Playlist
}

func (ctxt *context) onClientConnect(c *tcpserver.Client) {
	ctxt.log.Info("New client: ", c.RemoteAddr())
}

func (ctxt *context) onClientDisconnect(c *tcpserver.Client, err error) {
	ctxt.log.Warn("Client gone: ", c.RemoteAddr(), " because ", err)
}

func (ctxt *context) onNewMessage(c *tcpserver.Client, message []byte) {
	ctxt.log.Info("Msg: ", bytes.TrimRight(message, "\n"))
	ctxt.server.Broadcast(message)
}

func parseArgs(argv0 string) (args map[string]interface{}, err error) {
	usage := `{prog}.

Usage:
  {prog} [-c <configfile>]
  {prog} -h
  {prog} -v

Options:
  -c --config=<configfile>   Path to config file [default: config.toml]
  -h --help                  Show this screen.
  -v --version               Show version.`
	usage = strings.Replace(usage, "{prog}", argv0, -1)

	return docopt.Parse(usage, nil, true, "ury-listd-go "+LDVersion, false)
}

func main() {
	logger := logrus.New()

	// Get arguments (or their defaults)
	args, err := parseArgs(os.Args[0])
	if err != nil {
		logger.Fatal("Error parsing args: " + err.Error())
	}

	// Parse config
	// TODO: Make it its own type?
	var cfg struct {
		Server struct {
			Listen string
		}
		Playout struct {
			URI string
		}
		Log struct {
			Level string
		}
	}
	if _, err := toml.DecodeFile(args["--config"].(string), &cfg); err != nil {
		logger.Fatal("Error decoding toml config: " + err.Error())
	}

	// Properly set up logger
	if cfg.Log.Level != "" {
		level, err := logrus.ParseLevel(cfg.Log.Level)
		if err != nil {
			logger.Fatal("Failed to parse log level: " + err.Error())
		}
		logger.Level = level
	}

	// Make TCP Server
	s := tcpserver.New(cfg.Server.Listen)

	// TODO: Make connection to playout system

	c := &context{
		log:    logger,
		server: s,
	}

	s.SetClientConnectFunc(c.onClientConnect)
	s.SetClientDisconnectFunc(c.onClientDisconnect)
	s.SetNewMessageFunc(c.onNewMessage)

	c.playout, err = NewConnection(cfg.Playout.URI)

	logger.Info("Listening on ", cfg.Server.Listen)
	if err := s.Listen(); err != nil {
		logger.Error(err)
	}
}
