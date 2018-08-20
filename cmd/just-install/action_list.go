package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func handleListAction(c *cli.Context) {
	registry := loadRegistry(c)
	packageNames := registry.SortedPackageNames()

	for _, name := range packageNames {
		fmt.Printf("%35v - %v\n", name, registry.Packages[name].Version)
	}
}
