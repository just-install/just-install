package main

import (
	"log"

	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/justinstall"
)

func handleCleanAction(c *cli.Context) {
	if err := justinstall.CleanTempDir(); err != nil {
		log.Fatalln(err)
	}
}
