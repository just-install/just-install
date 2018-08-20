package main

import (
	"fmt"

	"github.com/codegangsta/cli"

	"github.com/just-install/just-install/pkg/justinstall"
)

func handleListAction(c *cli.Context) {
	registry := justinstall.SmartLoadRegistry(false)
	packageNames := registry.SortedPackageNames()

	for _, name := range packageNames {
		fmt.Printf("%35v - %v\n", name, registry.Packages[name].Version)
	}
}
