// just-install - The simple package installer for Windows
// Copyright (C) 2019 just-install authors.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"debug/pe"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/just-install/just-install/pkg/justinstall"
	"github.com/kardianos/osext"
	"github.com/urfave/cli"
)

var version = "## filled by go build ##"

func main() {
	app := cli.NewApp()
	app.Action = handleArguments
	app.Author = "just-install Developers"
	app.Name = "just-install"
	app.Usage = "The simple package installer for Windows"
	app.Version = version

	app.Commands = []cli.Command{{
		Name:   "audit",
		Usage:  "Audit the registry",
		Action: handleAuditAction,
	}, {
		Name:   "clean",
		Usage:  "Remove caches and temporary files",
		Action: handleCleanAction,
	}, {
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
		Name:  "download-only, d",
		Usage: "Only download packages, do not install them",
	}, cli.BoolFlag{
		Name:  "force, f",
		Usage: "Force package re-download",
	}, cli.StringFlag{
		Name:  "registry, r",
		Usage: "Use the specified registry file",
	}, cli.BoolFlag{
		Name:  "shim, s",
		Usage: "Create shims only (if exeproxy is installed)",
	}}

	// Extract arguments embedded in the executable (if any)
	pathname, err := osext.Executable()
	if err != nil {
		app.Run(os.Args)
		return
	}

	rawOverlayData, err := getPeOverlayData(pathname)
	if err != nil {
		app.Run(os.Args)
		return
	}

	stringOverlayData := string(rawOverlayData)
	trimmedStringOverlayData := strings.Trim(stringOverlayData, "\r\n ")
	if len(trimmedStringOverlayData) == 0 {
		app.Run(os.Args)
		return
	}

	log.Println("Using embedded arguments: " + trimmedStringOverlayData)
	app.Run(append([]string{os.Args[0]}, strings.Split(trimmedStringOverlayData, " ")...))
}

func handleArguments(c *cli.Context) {
	force := c.Bool("force")
	onlyDownload := c.Bool("download-only")
	onlyShims := c.Bool("shim")

	registry := loadRegistry(c)

	if c.String("arch") != "" {
		if err := justinstall.SetArchitecture(c.String("arch")); err != nil {
			log.Fatalln(err.Error())
		}
	}

	// Check which packages might require an interactive installation
	var interactive []string

	for _, pkg := range c.Args() {
		entry, ok := registry.Packages[pkg]
		if !ok {
			continue
		}

		if entry.Installer.Interactive {
			interactive = append(interactive, pkg)
		}
	}

	if len(interactive) > 0 {
		log.Println("These packages might require user interaction to complete their installation")

		for _, pkg := range interactive {
			log.Println("    " + pkg)
		}

		log.Println("")
	}

	// Install packages
	hasErrors := false

	for _, pkg := range c.Args() {
		entry, ok := registry.Packages[pkg]

		if ok {
			if onlyShims {
				entry.CreateShims()
			} else if onlyDownload {
				entry.DownloadInstaller(force)
			} else {
				if err := entry.JustInstall(force); err != nil {
					log.Printf("Error installing %v: %v", pkg, err)
					hasErrors = true
				}
			}
		} else {
			log.Println("WARNING: Unknown package", pkg)
		}
	}

	if hasErrors {
		log.Fatalln("Encountered errors installing packages")
	}
}

func getPeOverlayData(pathname string) ([]byte, error) {
	pefile, err := pe.Open(pathname)
	if err != nil {
		return nil, err
	}
	defer pefile.Close()

	lastSectionEnd := uint32(0)
	for v := range pefile.Sections {
		sectionHeader := pefile.Sections[v].SectionHeader
		sectionEnd := sectionHeader.Size + sectionHeader.Offset
		if sectionEnd > lastSectionEnd {
			lastSectionEnd = sectionEnd
		}
	}

	rawfile, err := ioutil.ReadFile(pathname)
	if err != nil {
		return nil, err
	}

	overlayRawData := rawfile[lastSectionEnd:]
	if len(overlayRawData) == 0 {
		return nil, errors.New("No overlay data found")
	}

	return overlayRawData, nil
}
