package main

import (
	"log"

	"github.com/codegangsta/cli"

	"github.com/just-install/just-install/pkg/justinstall"
)

func handleCleanAction(c *cli.Context) {
	if err := justinstall.CleanTempDir(); err != nil {
		log.Fatalln(err)
	}
}
