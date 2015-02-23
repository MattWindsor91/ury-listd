package main

import (
	"log"
	"os"

	"github.com/docopt/docopt-go"
)

func main() {
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

	arguments, _ := docopt.Parse(usage, nil, true, "ury-listd-go 0.0", false)
	logger := log.New(os.Stdout, "", log.Lshortfile)
	logger.Println(arguments)
}
