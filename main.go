// Command httpcheck runs HTTP health checks defined in YAML and reports the
// results to stdout (dev) or to a Sensu results API (prod), reading the check
// definitions from a local file (dev) or from Consul KV (prod).
package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/sergio-zerg/httpcheck/internal/engine"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Value:   "warning",
			Usage:   "logging level: debug, info, warning, error, fatal, panic",
			EnvVars: []string{"HTTPCHECK_LOG_LEVEL"},
		},
		&cli.StringFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Value:   "dev",
			Usage:   "dev: read checks from file, write to stdout. prod: read from Consul, write to Sensu",
			EnvVars: []string{"HTTPCHECK_ENV"},
		},
		&cli.StringFlag{
			Name:    "ip",
			Aliases: []string{"i"},
			Usage:   "override the Host header / source IP for every check",
			EnvVars: []string{"HTTPCHECK_IP"},
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"f"},
			Value:   "checks.yaml",
			Usage:   "path to the checks file (dev env)",
			EnvVars: []string{"HTTPCHECK_CONFIG"},
		},
		&cli.StringFlag{
			Name:    "consul-api",
			Aliases: []string{"c"},
			Value:   "http://consul-api:8500",
			Usage:   "Consul API base URL (prod env)",
			EnvVars: []string{"HTTPCHECK_CONSUL_API"},
		},
		&cli.StringFlag{
			Name:    "sensu-api",
			Aliases: []string{"s"},
			Value:   "http://sensu-api:4567",
			Usage:   "Sensu results API base URL (prod env)",
			EnvVars: []string{"HTTPCHECK_SENSU_API"},
		},
	}
}

func main() {
	app := &cli.App{
		Name:      "httpcheck",
		Usage:     "run HTTP health checks for web projects",
		Version:   version,
		Authors:   []*cli.Author{{Name: "sergio-zerg"}},
		Copyright: "MIT License",
		Flags:     flags(),
		Action: func(c *cli.Context) error {
			return engine.New(c).Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
