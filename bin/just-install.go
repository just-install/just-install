//
// just-install - The humble package installer for Windows
//
// Copyright (C) 2013, 2014, 2015, 2016, 2017 Lorenzo Villani
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
	"debug/pe"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/just-install/just-install"
	"github.com/kardianos/osext"
	dry "github.com/ungerik/go-dry"
)

var version = "## filled by go build ##"

func main() {
	app := cli.NewApp()
	app.Action = handleArguments
	app.Author = "Lorenzo Villani"
	app.Email = "lorenzo@villani.me"
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

	pathname, err := osext.Executable()
	if err != nil {
		app.Run(os.Args)
	} else {
		rawOverlayData, err := getPeOverlayData(pathname)
		if err != nil {
			app.Run(os.Args)
		} else {
			stringOverlayData := string(rawOverlayData)
			trimmedStringOverlayData := strings.Trim(stringOverlayData, "\r\n ")

			if len(trimmedStringOverlayData) == 0 {
				app.Run(os.Args)
			} else {
				log.Println("Using embedded arguments: " + trimmedStringOverlayData)
				app.Run(append([]string{os.Args[0]}, strings.Split(trimmedStringOverlayData, " ")...))
			}
		}
	}
}

func handleArguments(c *cli.Context) {
	force := c.Bool("force")
	onlyDownload := c.Bool("download-only")
	onlyShims := c.Bool("shim")

	var registry justinstall.Registry
	if c.IsSet("registry") {
		if !dry.FileExists(c.String("registry")) {
			log.Fatalf("%v: no such file.\n", c.String("registry"))
		}

		registry = justinstall.LoadRegistry(c.String("registry"))
	} else {
		registry = justinstall.SmartLoadRegistry(false)
	}

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
	for _, pkg := range c.Args() {
		entry, ok := registry.Packages[pkg]

		if ok {
			if onlyShims {
				entry.CreateShims()
			} else if onlyDownload {
				entry.DownloadInstaller(force)
			} else {
				entry.JustInstall(force)
			}
		} else {
			log.Println("WARNING: Unknown package", pkg)
		}
	}
}

func handleAuditAction(c *cli.Context) {
	expectedContentTypes := []string{
		"application/exe",
		"application/octet-stream",
		"application/unknown", // Bintray
		"application/x-dosexec",
		"application/x-msdos-program",
		"application/x-msdownload",
		"application/x-msi",
		"application/x-sdlc", // Oracle
		"application/x-zip-compressed",
		"application/zip",
		"binary/octet-stream",
		"Composite Document File V2 Document, corrupt: Can't read SAT; charset=binary", // Google Code
		"text/x-python", // PIP
		"Zip Files",
		"application/x-ms-dos-executable", // OpenVPN
		"exe", // VPN Unlimited
		"", // PIA
		"text/plain; charset=ISO-8859-1", // LibreOffice
	}

	// retry executes f, retrying a call with exponential back-off if it returns true as its first
	// return value. Ends up returning the eventual error value after a maximum of three retries.
	retry := func(f func() (bool, error)) error {
		var ret error

		delay := 2

		for i := 0; i < 3; i++ {
			shouldRetry, err := f()
			if !shouldRetry {
				return err
			}
			ret = err

			log.Println(err, "retrying in", delay, "seconds")
			time.Sleep(time.Duration(delay) * time.Second)
			delay *= delay
		}

		return ret
	}

	// FIXME: this chunk of code is duplicated with handleArguments().
	var registry justinstall.Registry
	if c.IsSet("registry") {
		if !dry.FileExists(c.String("registry")) {
			log.Fatalf("%v: no such file.\n", c.String("registry"))
		}

		registry = justinstall.LoadRegistry(c.String("registry"))
	} else {
		registry = justinstall.SmartLoadRegistry(false)
	}

	checkLink := func(rawurl string) error {
		return retry(func() (bool, error) {
			// Policy: retry on server or transport error, fail immediately otherwise.
			response, err := justinstall.CustomGet(rawurl)
			if err != nil {
				return true, err
			}
			defer response.Body.Close()

			if response.StatusCode >= 500 && response.StatusCode < 600 {
				return true, fmt.Errorf("%s: returned status code %v", rawurl, response.StatusCode)
			}

			if response.StatusCode != http.StatusOK {
				return false, fmt.Errorf("%s: expected status code 200, got %v", rawurl, response.StatusCode)
			}

			contentType := response.Header.Get("Content-Type")

			success := strings.HasSuffix(rawurl, ".vbox-extpack") && contentType == "text/plain"                     // VirtualBox Extension Pack has the wrong MIME type
			success = success || strings.Contains(rawurl, "libreoffice") && contentType == "application/x-troff-man" // Some LibreOffice mirrors return the wrong MIME type
			success = success || dry.StringInSlice(contentType, expectedContentTypes)
			if !success {
				return false, fmt.Errorf("%s: unexpected content type %q", rawurl, contentType)
			}

			return false, nil
		})
	}

	// Workers
	type workItem struct {
		description string
		rawurl      string
	}

	workerPoolSize := runtime.NumCPU()
	workerQueue := make(chan workItem, workerPoolSize)
	var workerWg sync.WaitGroup

	var collectedErrors []error
	for i := 0; i < workerPoolSize; i++ {
		workerWg.Add(1)

		go func() {
			for {
				item, more := <-workerQueue
				if !more {
					workerWg.Done()
					return
				}

				log.Println("checking", item.description)

				if err := checkLink(item.rawurl); err != nil {
					collectedErrors = append(collectedErrors, err)
				}
			}
		}()
	}

	// Push jobs to workers
	for _, name := range registry.SortedPackageNames() {
		entry := registry.Packages[name]

		if entry.Installer.X86 != "" {
			workerQueue <- workItem{name + " (x86)", entry.ExpandString(entry.Installer.X86)}
		}

		if entry.Installer.X86_64 != "" {
			workerQueue <- workItem{name + " (x86_64)", entry.ExpandString(entry.Installer.X86_64)}
		}
	}

	close(workerQueue)
	workerWg.Wait()

	if collectedErrors != nil {
		log.Println("Found errors:")

		for _, err := range collectedErrors {
			log.Println(err)
		}

		os.Exit(1)
	}
}

func handleCleanAction(c *cli.Context) {
	if err := justinstall.CleanTempDir(); err != nil {
		log.Fatalln(err)
	}
}

func handleListAction(c *cli.Context) {
	registry := justinstall.SmartLoadRegistry(false)
	packageNames := registry.SortedPackageNames()

	for _, name := range packageNames {
		fmt.Printf("%35v - %v\n", name, registry.Packages[name].Version)
	}
}

func handleUpdateAction(c *cli.Context) {
	justinstall.SmartLoadRegistry(true)
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
