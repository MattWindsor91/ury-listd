package main

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
)

// Version string, provided by linker flags
var LDVersion string

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

	// Make Context
	c := NewContext(cfg.Server.Listen, cfg.Playout.URI, logger)

	c.Run()
}
