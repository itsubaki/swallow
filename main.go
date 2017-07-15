package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "swallow"
	app.Usage = "hipchat client."
	app.Version = "0.0.1"
	app.Run(os.Args)
}
