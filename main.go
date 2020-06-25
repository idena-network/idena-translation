package main

import (
	"github.com/idena-network/idena-translation/config"
	"github.com/idena-network/idena-translation/types"
	"gopkg.in/urfave/cli.v1"
	"os"
)

// @license.name Apache 2.0
func main() {
	app := cli.NewApp()
	app.Name = "github.com/idena-network/idena-translation"
	app.Version = types.AppVersion

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "Config file",
			Value: "config.json",
		},
	}
	app.Action = func(context *cli.Context) error {
		appConfig := config.LoadConfig(context.String("config"))
		startServer(appConfig)
		return nil
	}
	app.Run(os.Args)
}
