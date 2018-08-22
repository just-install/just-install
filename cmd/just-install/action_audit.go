package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ungerik/go-dry"
	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/justinstall"
)

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
		"exe",                             // VPN Unlimited
		"",                                // PIA
		"text/plain; charset=ISO-8859-1",  // LibreOffice
		"text/html; charset=utf-8",        // SourceForge
		"application/x-ole-storage",       // EpicGames
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

	registry := loadRegistry(c)

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
				if strings.Contains(response.Request.URL.Host, "freefilesync") || strings.Contains(response.Request.URL.Host, "mediafire") {
					// mediafire is unpredictable
					return true, nil
				}
				if strings.Contains(response.Request.URL.Host, "download.gimp.org") {
					// gimp does funny stuff when trying to download from appveyor, but it works properly when actually using j-i
					return true, nil
				}
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
