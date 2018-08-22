package main

import (
	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/justinstall"
)

func handleUpdateAction(c *cli.Context) {
	justinstall.SmartLoadRegistry(true)
}
