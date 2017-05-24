package main

import (
	"github.com/urfave/cli"
	"github.com/sergio-zerg/httpcheck/engine"
	"os"
	"sort"
	"time"
)

const version = "1.0.0"

func flags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:   "log-level, l",
			Value:  "warning",
			Usage:  "Set logging level: info, warning, error, fatal, debug, panic",
			EnvVar: "HTTPCHECK_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "env, e",
			Value:  "dev",
			Usage:  "dev: input - set with flag -c (default config.yaml), output - stdout. prod: input - consul, output - sensu",
			EnvVar: "HTTPCHECK_ENV",
		},
		cli.StringFlag{
			Name: "ip, i",
			//Value:  "127.0.0.1",
			Usage:  "ip-address for checks",
			EnvVar: "HTTPCHECK_IP",
		},
		cli.StringFlag{
			Name:   "config, f",
			Value:  "config.yaml",
			Usage:  "ip-address for checks",
			EnvVar: "HTTPCHECK_CONFIG",
		},
		cli.StringFlag{
			Name:   "consul-api, c",
			Value:  "http://consul-api:8500",
			Usage:  "Consul API host",
			EnvVar: "HTTPCHECK_CONSUL_API",
		},
		cli.StringFlag{
			Name:   "sensu-api, s",
			Value:  "http://sensu-api:4567",
			Usage:  "Sensu API host",
			EnvVar: "HTTPCHECK_SENSU_API",
		},
		//@todo write validator for this flag?
		cli.StringFlag{
			Name:  "validate, d",
			Usage: "validate config",
		},
	}
	return flags
}

func main() {
	app := cli.NewApp()
	app.Compiled = time.Now()
	app.Name = "httpcheck"
	app.Usage = "make check web projects"
	app.Author = "sergio_zerg"
	app.Email = "zerg123@gmail.com"
	app.Copyright = "MIT License"
	app.HelpName = "httpcheck"
	app.Version = version
	app.Flags = flags()
	app.Action = func(c *cli.Context) {
		core := engine.NewEngine(c)
		core.Run()
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}
