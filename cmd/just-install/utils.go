package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/ungerik/go-dry"

	"github.com/just-install/just-install/pkg/justinstall"
)

func loadRegistry(c *cli.Context) justinstall.Registry {
	if !c.GlobalIsSet("registry") {
		return justinstall.SmartLoadRegistry(false)
	}

	registryPath := c.GlobalString("registry")
	if !dry.FileExists(registryPath) {
		log.Fatalf("%v: no such file.\n", registryPath)
	}

	log.Println("Loading custom registry at", registryPath)
	return justinstall.LoadRegistry(registryPath)
}
