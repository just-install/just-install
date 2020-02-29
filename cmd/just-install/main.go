// just-install - The simple package installer for Windows
// Copyright (C) 2020 just-install authors.
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

	"github.com/urfave/cli/v2"

	"github.com/just-install/just-install/pkg/platform"
)

var version = "## filled by go build ##"

func main() {
	app := cli.NewApp()
	app.Action = handleArguments
	app.Name = "just-install"
	app.Usage = "The simple package installer for Windows"
	app.Version = version

	app.Commands = []*cli.Command{{
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

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Aliases: []string{"a"},
			Name:    "arch",
			Usage:   "Force installation for a specific architecture (if supported by the host).",
		}, &cli.BoolFlag{
			Aliases: []string{"d"},
			Name:    "download-only",
			Usage:   "Only download packages, do not install them",
		}, &cli.BoolFlag{
			Aliases: []string{"f"},
			Name:    "force",
			Usage:   "Force package re-download",
		}, &cli.StringFlag{
			Aliases: []string{"r"},
			Name:    "registry",
			Usage:   "Use the specified registry file",
		}, &cli.BoolFlag{
			Aliases: []string{"s"},
			Name:    "shim",
			Usage:   "Create shims only (if exeproxy is installed)",
		},
	}

	// Normalize "%ProgramFiles%" and "%ProgramFiles(x86)%"
	platform.SetNormalisedProgramFilesEnv()

	// Extract arguments embedded in the executable (if any)
	pathname, err := os.Executable()
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
	if err := app.Run(append([]string{os.Args[0]}, strings.Split(trimmedStringOverlayData, " ")...)); err != nil {
		log.Fatalln(err)
	}
}

func handleArguments(c *cli.Context) error {
	force := c.Bool("force")
	onlyDownload := c.Bool("download-only")
	onlyShims := c.Bool("shim")

	registry, err := loadRegistry(c, force)
	if err != nil {
		return err
	}

	// Architecture selection
	arch := c.String("arch")
	switch arch {
	case "":
		if platform.Is64Bit() {
			arch = "x86_64"
		} else {
			arch = "x86"
		}
	case "x86":
		// Nothing to do
	case "x86_64":
		if !platform.Is64Bit() {
			log.Fatalln("This machine cannot run 64-bit software")
		}
	default:
		log.Fatalln("Unknown architecture:", arch)
	}

	// Check which packages might require an interactive installation
	var interactive []string

	for _, pkg := range c.Args().Slice() {
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

	for _, pkg := range c.Args().Slice() {
		entry, ok := registry.Packages[pkg]

		if ok {
			if onlyShims {
				entry.CreateShims(arch)
			} else if onlyDownload {
				entry.DownloadInstaller(arch, force)
			} else {
				if err := entry.JustInstall(arch, force); err != nil {
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

	return nil
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
