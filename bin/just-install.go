//
// just-install - The stupid package installer
//
// Copyright (C) 2013, 2014, 2015  Lorenzo Villani
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.	 See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/lvillani/just-install"
)

const (
	version = "3.0.0"
)

func main() {
	app := cli.NewApp()
	app.Author = "Lorenzo Villani"
	app.Email = "lorenzo@villani.me"
	app.Name = "just-install"
	app.Usage = "The stupid package installer for Windows"
	app.Version = version
	app.Action = handleArguments

	app.Commands = []cli.Command{{
		Name:   "list",
		Usage:  "List all known packages",
		Action: handleListAction,
	}, {
		Name:   "update",
		Usage:  "Update the registry",
		Action: handleUpdateAction,
	}}

	app.Flags = []cli.Flag{cli.StringFlag{
		Name:  "arch, a",
		Usage: "Force installation for a specific architecture (if supported by the host).",
	}, cli.BoolFlag{
		Name:  "force, f",
		Usage: "Force package re-download",
	}, cli.BoolFlag{
		Name:  "shim, s",
		Usage: "Create shims only (if exeproxy is installed)",
	}}

	app.Run(os.Args)
}

func handleArguments(c *cli.Context) {
	force := c.Bool("force")
	onlyShims := c.Bool("shim")
	registry := justinstall.SmartLoadRegistry(false)

	if c.String("arch") != "" {
		if err := justinstall.SetArchitecture(c.String("arch")); err != nil {
			log.Fatalln(err.Error())
		}
	}

	// Install packages
	for _, pkg := range c.Args() {
		entry, ok := registry.Packages[pkg]

		if ok {
			if onlyShims {
				entry.CreateShims()
			} else {
				entry.JustInstall(force)
			}
		} else {
			log.Println("WARNING: Unknown package", pkg)
		}
	}
}

func handleListAction(c *cli.Context) {
	registry := justinstall.SmartLoadRegistry(false)

	fmt.Println(strings.Join(registry.SortedPackageNames(), "\n"))
}

func handleUpdateAction(c *cli.Context) {
	justinstall.SmartLoadRegistry(true)
}
