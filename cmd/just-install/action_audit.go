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
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/just-install/just-install/pkg/fetch"
)

func handleAuditAction(c *cli.Context) error {
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
		"exe",                             // VPN Unlimited
		"",                                // PIA
		"text/plain; charset=ISO-8859-1",  // LibreOffice
		"text/html; charset=utf-8",        // SourceForge
		"application/x-ole-storage",       // EpicGames
		"application/x-troff-man",         // MSIs on OSDN
		"application/x-executable",        // Notepad++
	}

	// retry executes f, retrying a call with exponential back-off if it returns an error and true
	// as its first return value. Ends up returning the eventual error value after a maximum of
	// three retries.
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

	registry, err := loadRegistry(c, c.Bool("force"), !c.Bool("noprogress"))
	if err != nil {
		return err
	}

	checkLink := func(rawurl string) error {
		return retry(func() (bool, error) {
			// Policy: retry on server or transport error, fail immediately otherwise.
			err := fetch.Check(rawurl, &fetch.CheckOptions{ExpectedContentTypes: expectedContentTypes})
			if _, ok := err.(*fetch.HTTPStatusError); ok {
				return false, err
			}

			return err != nil, err
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
		if entry.SkipAudit {
			log.Println("skipping audit of", name)
			continue
		}

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
		log.Println("found errors:")

		for _, err := range collectedErrors {
			log.Println(err)
		}

		os.Exit(1)
	}

	return nil
}
