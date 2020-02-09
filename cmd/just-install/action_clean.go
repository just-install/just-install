package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/paths"
)

func handleCleanAction(c *cli.Context) {
	// Yup, this is weird, but we don't want a public API that allows us to use the temporary
	// directory before creating it elsewhere in the program.
	tempDir, err := paths.TempDirCreate()
	if err != nil {
		log.Fatalln("Could not create temporary directory:", err)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		log.Fatalln("Could not clean temporary directory:", err)
	}
}
