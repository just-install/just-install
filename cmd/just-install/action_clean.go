package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/paths"
)

func handleCleanAction(c *cli.Context) {
	if err := os.RemoveAll(paths.TempDir()); err != nil {
		log.Fatalln(err)
	}

	if _, err := paths.TempDirCreate(); err != nil {
		log.Fatalln(err)
	}
}
