package main

import (
	"github.com/urfave/cli"
)

func handleUpdateAction(c *cli.Context) {
	if err := c.GlobalSet("force", "true"); err != nil {
		panic(err)
	}

	loadRegistry(c)
}
