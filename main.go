package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	baps3 "github.com/UniversityRadioYork/baps3-go"
	"github.com/docopt/docopt-go"
)

var LD_VERSION string

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

	return docopt.Parse(usage, nil, true, "ury-listd-go 0.0", false)
}

func main() {
	log.SetFlags(log.Lshortfile) // Set up default logger
	args, err := parseArgs()
	if err != nil {
		log.Fatal("Error parsing args: " + err.Error())
	}

	// Set up connection to playout
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT)

	responseCh := make(chan baps3.Message)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	connLog := log.New(os.Stderr, "playd:", 0)
	connector := baps3.InitConnector("", responseCh, wg, connLog)
	connector.Connect(args["--playoutaddr"].(string) + ":" + args["--playoutport"].(string))
	go connector.Run()

	h.setConnector(connector.ReqCh, responseCh)

	go h.runListener(args["--addr"].(string), args["--port"].(string))

	// Signal handler loop
	for {
		select {
		case <-sigs:
			log.Println("Exiting...")
			h.Quit <- true
			//<-h.Quit // Wait for quit to finish
			close(connector.ReqCh)
			wg.Wait()
			os.Exit(0)
		}
	}
}
