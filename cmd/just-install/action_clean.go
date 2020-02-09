package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/paths"
)

func handleCleanAction(c *cli.Context) {
	// NOTE: the temporary directory will be recreated at the next run from main()
	if err := os.RemoveAll(paths.TempDir()); err != nil {
		log.Fatalln(err)
	}
}
