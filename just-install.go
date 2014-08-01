//
// just-install - The stupid package installer
//
// Copyright (C) 2013, 2014  Lorenzo Villani
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
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/kardianos/osext"
	"github.com/codegangsta/cli"
	"github.com/inconshreveable/go-update"
	"gopkg.in/cheggaaa/pb.v0"
)

const (
	REGISTRY        = "https://raw.github.com/lvillani/just-install/master/just-install.json"
	SELF_UPDATE_URL = "https://github.com/lvillani/just-install/raw/gh-pages/just-install.exe"
)

var (
	registryPath    = filepath.Join(os.TempDir(), "just-install.json")
	selfInstallPath = filepath.Join(os.Getenv("WINDIR"), "just-install.exe")
)

//
// Registry Schema
//

type Registry struct {
	Version  int
	Packages map[string]RegistryEntry
}

type RegistryEntry struct {
	Version   string
	Installer InstallerEntry
}

type InstallerEntry struct {
	Kind    string
	X86     string
	X86_64  string
	Options map[string]string
}

func (e *RegistryEntry) JustInstall(force bool, arch string) {
	url := e.pickInstallerUrl(arch)
	url = strings.Replace(url, "${version}", e.Version, -1)

	log.Println(arch, "-", url)

	e.install(download2(url, force))
}

func (e *RegistryEntry) pickInstallerUrl(arch string) string {
	if arch == "x86_64" && isAmd64() && e.Installer.X86_64 != "" {
		return e.Installer.X86_64
	} else {
		return e.Installer.X86
	}
}

// Returns `true` if the host system is 64-bit capable, `false` otherwise.
func isAmd64() bool {
	return 1<<32 != 0
}

func (e *RegistryEntry) install(installer string) {
	if e.Installer.Kind == "advancedinstaller" {
		e.exec(installer, "/q", "/i")
	} else if e.Installer.Kind == "as-is" {
		e.exec(installer)
	} else if e.Installer.Kind == "conemu" {
		if isAmd64() {
			e.exec(installer, "/p:x64", "/q")
		} else {
			e.exec(installer, "/p:x86", "/q")
		}
	} else if e.Installer.Kind == "easy_install_26" {
		e.exec("\\Python26\\Scripts\\easy_install.exe", installer)
	} else if e.Installer.Kind == "easy_install_27" {
		e.exec("\\Python27\\Scripts\\easy_install.exe", installer)
	} else if e.Installer.Kind == "innosetup" {
		e.exec(installer, "/norestart", "/sp-", "/verysilent")
	} else if e.Installer.Kind == "msi" {
		e.exec("msiexec.exe", "/q", "/i", installer, "REBOOT=ReallySuppress")
	} else if e.Installer.Kind == "nsis" {
		e.exec(installer, "/S", "/NCRC")
	} else {
		log.Fatalln("Unknown installer type:", e.Installer.Kind)
	}
}

func (e *RegistryEntry) exec(installer string, args ...string) {
	for i, a := range args {
		args[i] = strings.Replace(a, "${installer}", installer, -1)
	}

	log.Println("Running", installer, args)

	cmd := exec.Command(os.Getenv("COMSPEC"), append([]string{"/C", installer}, args...)...)
	err := cmd.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

//
// Funcs
//

func main() {
	app := cli.NewApp()
	app.Author = "Lorenzo Villani"
	app.Email = "lorenzo@villani.me"
	app.Name = "just-install"
	app.Usage = "The stupid package installer for Windows"
	app.Version = "2.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "arch, a",
			Usage: "Force installation for a specific architecture (if supported by the host).",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Force package re-download",
		},
	}
	app.Action = func(c *cli.Context) {
		selfInstall()

		force := c.Bool("force")
		registry := smartLoadRegistry(force)

		// Architecture selection
		arch := c.String("arch")

		if arch == "" {
			if isAmd64() {
				arch = "x86_64"
			} else {
				arch = "x86"
			}
		} else if arch == "x86_64" && !isAmd64() {
			log.Fatalln("Your machine is not 64-bit capable.")
		} else if arch != "x86" && arch != "x86_64" {
			log.Fatalln("Please specify a valid architecture between x86 and x86_64")
		}

		// Install packages
		for _, pkg := range c.Args() {
			entry := registry.Packages[pkg]
			entry.JustInstall(force, arch)
		}
	}
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "List all known packages",
			Action: func(c *cli.Context) {
				registry := smartLoadRegistry(false)

				for k, v := range registry.Packages {
					fmt.Printf("%s: %s\n", k, v.Version)
				}
			},
		},
		{
			Name:  "self-update",
			Usage: "Update just-install itself",
			Action: func(c *cli.Context) {
				log.Println("Self-updating...")

				update := update.New()

				err, _ := update.FromUrl(SELF_UPDATE_URL)
				if err != nil {
					log.Fatalln(err.Error())
				}
			},
		},
		{
			Name:  "update",
			Usage: "Update the registry",
			Action: func(c *cli.Context) {
				smartLoadRegistry(true)
			},
		}}
	app.Run(os.Args)
}

// Copy ourselves to %WINDIR%\just-install.exe in case we are not being executed from there.
func selfInstall() {
	executable, err := osext.Executable()
	if err != nil {
		log.Println("Unable to determine where I'm running from. Cannot self-install.")

		return
	}

	if executable != selfInstallPath {
		log.Println("Self installing to:", selfInstallPath)

		copyFile(os.Args[0], selfInstallPath)
	}
}

func copyFile(src string, dst string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return ioutil.WriteFile(dst, buf, 0)
}

// Loads the development registry, if there. Otherwise tries to load a cached copy downloaded from
// the Internet (downloading it if missing).
func smartLoadRegistry(force bool) Registry {
	if fileExists("just-install.json") {
		log.Println("Using local registry file")

		return loadRegistry("just-install.json")
	} else {
		if !fileExists(registryPath) || force {
			log.Println("Updating registry from:", REGISTRY)

			downloadRegistry()
		}

		return loadRegistry(registryPath)
	}
}

// Returns `true` if there is a file at the given `path`. Returns `false` otherwise.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// Downloads the registry from the canonical URL.
func downloadRegistry() {
	download(REGISTRY, registryPath)
}

// Unmarshals the registry from a local file path.
func loadRegistry(path string) Registry {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to parse the registry file.")
	}

	var ret Registry

	json.Unmarshal(data, &ret)

	return ret
}

// Downloads a file over HTTP(S) to a temporary location. The temporary file has a name derived
// from the CRC32 of the URL string with the original file extension attached (if any). The file
// is re-downloaded only if the temporary file is missing or `force` is true.
func download2(rawurl string, force bool) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatalf("Unable to parse the URL: %s", rawurl)
	}

	base := crc32s(rawurl) + filepath.Ext(u.Path)
	dest := filepath.Join(os.TempDir(), base)

	if !fileExists(dest) || force {
		download(rawurl, dest)
	}

	return dest
}

// Computes and returns the CRC32 of a string as an HEX string.
func crc32s(s string) string {
	crc32 := crc32.NewIEEE()
	crc32.Write([]byte(s))

	return fmt.Sprintf("%X", crc32.Sum32())
}

// Downloads a file with the HTTP/HTTPS protocol showing a progress bar. The destination file is
// always overwritten.
func download(rawurl string, destinationPath string) {
	destination, err := os.Create(destinationPath)
	if err != nil {
		log.Fatalf("Unable to open the destination file: %s", destinationPath)
	}

	defer destination.Close()

	response, err := http.Get(rawurl)
	if err != nil {
		log.Fatalf("Unable to open a connection to %s", rawurl)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected HTTP response code. Wanted 200 but got %d", response.StatusCode)
	}

	var progressBar *pb.ProgressBar

	contentLength, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err == nil {
		progressBar = pb.New(int(contentLength))
	} else {
		progressBar = pb.New(0)
	}

	progressBar.ShowSpeed = true
	progressBar.SetRefreshRate(time.Millisecond * 1000)
	progressBar.SetUnits(pb.U_BYTES)
	progressBar.Start()

	writer := io.MultiWriter(destination, progressBar)

	io.Copy(writer, response.Body)

	// Cleanup
	progressBar.Finish()
	destination.Close()
	response.Body.Close()
}
