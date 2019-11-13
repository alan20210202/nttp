package main

import (
	"github.com/urfave/cli"
	"log"
	"nttp"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "nttp"
	app.Usage = "NTT-coated SOCKS client/server"
	app.Version = "998.244.353"
	app.Commands = []*cli.Command{
		{
			Name:  "client",
			Usage: "Run nttp as client",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "listen, l",
					Usage:    "address on which nttp listens for incoming SOCKS requests",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "remote, r",
					Usage:    "remote address",
					Required: true,
				},
			},
			Action: func(ctx *cli.Context) error {
				nttp.ListenAsClient(ctx.String("listen"), ctx.String("remote"))
				return nil
			},
		},
		{
			Name:  "server",
			Usage: "Run nttp as server",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "listen, l",
					Usage:    "address on which nttp listens for clients",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "self, s",
					Value: "0.0.0.0",
					Usage: "public address of the server",
				},
			},
			Action: func(ctx *cli.Context) error {
				nttp.ListenAsServer(ctx.String("listen"), ctx.String("self"))
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
