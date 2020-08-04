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
	app.Action = handleInstall
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
		}, &cli.BoolFlag{
			Aliases: []string{"no-progress"},
			Name:    "noprogress",
			Usage:   "Don't display progress bar",
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
		if err := app.Run(os.Args); err != nil {
			log.Fatalln(err)
		}

		return
	}

	rawOverlayData, err := getPeOverlayData(pathname)
	if err != nil {
		if err := app.Run(os.Args); err != nil {
			log.Fatalln(err)
		}

		return
	}

	stringOverlayData := string(rawOverlayData)
	trimmedStringOverlayData := strings.Trim(stringOverlayData, "\r\n ")
	if len(trimmedStringOverlayData) == 0 {
		if err := app.Run(os.Args); err != nil {
			log.Fatalln(err)
		}

		return
	}

	log.Println("using embedded arguments: " + trimmedStringOverlayData)
	if err := app.Run(append([]string{os.Args[0]}, strings.Split(trimmedStringOverlayData, " ")...)); err != nil {
		log.Fatalln(err)
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
		return nil, errors.New("no overlay data found")
	}

	return overlayRawData, nil
}
