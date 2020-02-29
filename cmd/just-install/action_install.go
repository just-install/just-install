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
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/just-install/just-install/pkg/platform"
)

func handleInstall(c *cli.Context) error {
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
			return errors.New("this machine cannot run 64-bit software")
		}
	default:
		return fmt.Errorf("unknown architecture: %v", arch)
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
		log.Println("these packages might require user interaction to complete their installation")

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
					log.Printf("error installing %v: %v", pkg, err)
					hasErrors = true
				}
			}
		} else {
			log.Println("WARNING: unknown package", pkg)
		}
	}

	if hasErrors {
		return errors.New("encountered errors installing packages")
	}

	return nil
}
