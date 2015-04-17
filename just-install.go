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
	"archive/zip"
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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ungerik/go-dry"
	"gopkg.in/cheggaaa/pb.v0"
)

const (
	registrySupportedVersion = 2
	registryURL              = "https://raw.github.com/lvillani/just-install/v3.0/just-install.json"
	version                  = "3.0.0"
)

var (
	arch         = "x86"
	isAmd64      = false
	registryPath = filepath.Join(os.TempDir(), "just-install.json")
	shimsPath    = os.ExpandEnv("${SystemDrive}\\just-install")
)

func main() {
	determineArch()
	normalizeProgramFiles()
	handleCommandLine()
}

// determineArch determines the Windows architecture of the current Windows installation. It changes
// both the "isAmd64" and "arch" globals.
func determineArch() {
	// Since our output is a 32-bit executable (for maximum compatibility) and all other options
	// proved fruitless, let's just test for something that is usually available only on x86_64
	// editions of Windows.
	sentinel := os.Getenv("ProgramFiles(x86)")

	if sentinel == "" {
		isAmd64 = false
	} else {
		isAmd64 = dry.FileIsDir(sentinel)
	}

	arch = registryArch()
}

// normalizeProgramFiles re-exports environment variables so that %ProgramFiles% and
// %ProgramFiles(x86)% always point to the same directory on 32-bit systems and %ProgramFiles%
// points to the 64-bit directory even if we are a 32-bit binary.
func normalizeProgramFiles() {
	// Disabling SysWOW64 is a bad idea and going with Win32 API proved fruitless.
	// Time to get dirty.
	var programFiles string
	var programFilesX86 string

	if isAmd64 {
		programFilesX86 = os.Getenv("ProgramFiles(x86)")
		programFiles = programFilesX86[0:strings.LastIndex(programFilesX86, " (x86)")]
	} else {
		programFiles = os.Getenv("ProgramFiles")
		programFilesX86 = programFiles
	}

	os.Setenv("ProgramFiles", programFiles)
	os.Setenv("ProgramFiles(x86)", programFilesX86)
}

func handleCommandLine() {
	app := cli.NewApp()
	app.Author = "Lorenzo Villani"
	app.Email = "lorenzo@villani.me"
	app.Name = "just-install"
	app.Usage = "The stupid package installer for Windows"
	app.Version = version
	app.Action = handleArguments

	app.Commands = []cli.Command{{
		Name:   "checklinks",
		Usage:  "Check for bad installer links",
		Action: handleChecklinksAction,
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
	registry := smartLoadRegistry(false)

	if c.String("arch") != "" {
		arch = preferredArch(c.String("arch"))
	}

	// Install packages
	for _, pkg := range c.Args() {
		entry, ok := registry.Packages[pkg]

		if ok {
			if onlyShims {
				entry.createShims()
			} else {
				entry.JustInstall(force, arch)
			}
		} else {
			log.Println("WARNING: Unknown package", pkg)
		}
	}
}

func handleChecklinksAction(c *cli.Context) {
	checkLinks(smartLoadRegistry(false))
}

func handleListAction(c *cli.Context) {
	registry := smartLoadRegistry(false)

	for _, v := range sortedKeys(registry.Packages) {
		fmt.Printf("%s: %s\n", v, registry.Packages[v].Version)
	}
}

func handleUpdateAction(c *cli.Context) {
	smartLoadRegistry(true)
}

// Loads the development registry, if there. Otherwise tries to load a cached copy downloaded from
// the Internet. If neither is available, try to download it from the known location first.
func smartLoadRegistry(force bool) registry {
	if dry.FileExists("just-install.json") {
		log.Println("Using local registry file")

		return loadRegistry("just-install.json")
	}

	if !dry.FileExists(registryPath) || force {
		log.Println("Updating registry from:", registryURL)

		downloadRegistry()
	}

	return loadRegistry(registryPath)
}

// Unmarshals the registry from a local file path.
func loadRegistry(path string) registry {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read the registry file.")
	}

	var ret registry

	if err := json.Unmarshal(data, &ret); err != nil {
		log.Fatalln("Unable to parse the registry file.")
	}

	if ret.Version != registrySupportedVersion {
		log.Fatalln("Please update to a new version of just-install by running: msiexec.exe /i http://go.just-install.it")
	}

	return ret
}

// Downloads the registry from the canonical URL.
func downloadRegistry() {
	download(registryURL, registryPath)
}

//
// Installer Entry
//

type installerEntry struct {
	Container string // Optional
	Kind      string
	X86       string
	X86_64    string
	Options   map[string]interface{} // Optional
}

// options returns the architecture-specific options (if available), otherwise returns the whole
// options map.
func (s *installerEntry) options() map[string]interface{} {
	archSpecificOptions, ok := s.Options[arch].(map[string]interface{})
	if !ok {
		return s.Options
	}

	return archSpecificOptions
}

//
// Registry
//

type registry struct {
	Version  int
	Packages map[string]registryEntry
}

type registryEntry struct {
	Version   string
	Installer installerEntry
}

func (e *registryEntry) JustInstall(force bool, arch string) {
	url := e.pickInstallerURL(arch)
	url = strings.Replace(url, "${version}", e.Version, -1)

	log.Println(arch, "-", url)

	var downloadedFile string

	if ext, ok := e.Installer.options()["extension"]; ok {
		downloadedFile = download3(url, ext.(string), force)
	} else {
		downloadedFile = download2(url, force)
	}

	if e.Installer.Container != "" {
		// We first need to unwrap the container, then read the real file name to install
		// from `Options` and run it.
		tempDir := e.unwrap(downloadedFile, e.Installer.Container)
		install, ok := e.Installer.options()["install"].(string)

		if !ok {
			log.Fatalln("Specified a container but wasn't told where is the real installer.")
		}

		e.install(filepath.Join(tempDir, install))
	} else {
		// Run the installer as-is
		e.install(downloadedFile)
	}

	e.createShims()
}

func (e *registryEntry) pickInstallerURL(arch string) string {
	if arch == "x86_64" && isAmd64 && e.Installer.X86_64 != "" {
		return e.Installer.X86_64
	}

	return e.Installer.X86
}

// Extracts the given container file to a temporary directory and returns that paths.
func (e *registryEntry) unwrap(containerPath string, kind string) string {
	if kind == "zip" {
		extractTo := filepath.Join(os.TempDir(), crc32s(containerPath))

		extractZip(containerPath, extractTo)

		return extractTo
	}

	log.Fatalln("Unknown container type:", kind)

	return "" // We should never get here.
}

func (e *registryEntry) install(installer string) {
	if e.Installer.Kind == "advancedinstaller" {
		system(installer, "/q", "/i")
	} else if e.Installer.Kind == "as-is" {
		system(installer)
	} else if e.Installer.Kind == "custom" {
		var args []string

		for _, v := range e.Installer.options()["arguments"].([]interface{}) {
			current := strings.Replace(v.(string), "${installer}", installer, -1)
			current = os.ExpandEnv(current)

			args = append(args, current)
		}

		if len(args) == 0 {
			return
		} else if len(args) == 1 {
			system(args[0])
		} else {
			system(args[0], args[1:]...)
		}
	} else if e.Installer.Kind == "easy_install_26" {
		system("\\Python26\\Scripts\\easy_install.exe", installer)
	} else if e.Installer.Kind == "easy_install_27" {
		system("\\Python27\\Scripts\\easy_install.exe", installer)
	} else if e.Installer.Kind == "innosetup" {
		system(installer, "/norestart", "/sp-", "/verysilent")
	} else if e.Installer.Kind == "msi" {
		system("msiexec.exe", "/q", "/i", installer, "ALLUSERS=1", "REBOOT=ReallySuppress")
	} else if e.Installer.Kind == "nsis" {
		system(installer, "/S", "/NCRC")
	} else if e.Installer.Kind == "zip" {
		destination := os.ExpandEnv(e.Installer.options()["destination"].(string))

		log.Println("Extracting to", destination)

		extractZip(installer, os.ExpandEnv(e.Installer.options()["destination"].(string)))
	} else {
		log.Fatalln("Unknown installer type:", e.Installer.Kind)
	}
}

func (e *registryEntry) createShims() {
	exeproxy := os.ExpandEnv("${ProgramFiles(x86)}\\exeproxy\\exeproxy.exe")

	if !dry.FileExists(exeproxy) {
		return
	}

	if !dry.FileIsDir(shimsPath) {
		os.MkdirAll(shimsPath, 0)
	}

	if shims, ok := e.Installer.options()["shims"]; ok {
		for _, v := range shims.([]interface{}) {
			shimTarget := strings.Replace(v.(string), "${version}", e.Version, -1)
			shimTarget = os.ExpandEnv(shimTarget)
			shim := filepath.Join(shimsPath, filepath.Base(shimTarget))

			if dry.FileExists(shim) {
				os.Remove(shim)
			}

			log.Printf("Creating shim for %s (%s)\n", shimTarget, shim)

			system(exeproxy, "exeproxy-copy", shim, shimTarget)
		}
	}
}

//
// Utilities
//

// preferredArch returns the given architecture if it is valid and supported by the system.
// Otherwise it returns the name for the current architecture (see `registryArch`). Please note that
// this function terminates the application if the preferred architecture is either invalid or not
// supported.
func preferredArch(arch string) string {
	if arch == "x86_64" && !isAmd64 {
		log.Fatalln("Your machine is not 64-bit capable")
	} else if arch != "x86" && arch != "x86_64" {
		log.Fatalln("Please specify a valid architecture between x86 and x86_64")
	} else if arch == "" {
		return registryArch()
	}

	return arch
}

// registryArch returns a string which represents the current architecture in the registry file.
func registryArch() string {
	if isAmd64 {
		return "x86_64"
	}

	return "x86"
}

func system(command string, args ...string) {
	log.Println("Running", command, args)

	cmd := exec.Command(command, args...)
	err := cmd.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func sortedKeys(m map[string]registryEntry) []string {
	keys := make([]string, len(m))
	i := 0

	for k := range m {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

func copyFile(src string, dst string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return ioutil.WriteFile(dst, buf, 0)
}

// Convenience wrapper over download3 which passes an empty ("") `ext` parameter.
func download2(rawurl string, force bool) string {
	return download3(rawurl, "", force)
}

// Downloads a file over HTTP(S) to a temporary location. The temporary file has a name derived
// from the CRC32 of the URL string with the original file extension attached (if any). If `ext`
// is not the empty string, it will be appended to the destination file. The file is re-downloaded
// only if the temporary file is missing or `force` is true.
func download3(rawurl string, ext string, force bool) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatalf("Unable to parse the URL: %s", rawurl)
	}

	var base string

	if ext != "" {
		base = crc32s(rawurl) + ext
	} else {
		base = crc32s(rawurl) + filepath.Ext(u.Path)
	}

	dest := filepath.Join(os.TempDir(), base)

	if !dry.FileExists(dest) || force {
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

func extractZip(path string, extractTo string) {
	os.MkdirAll(extractTo, 0700)

	// Open the archive for reading
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		log.Fatalln("Unable to open ZIP archive:", path)
	}
	defer zipReader.Close()

	// Extract all entries in the archive
	for _, zipFile := range zipReader.File {
		destinationPath := filepath.Join(extractTo, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			os.MkdirAll(destinationPath, zipFile.Mode())
		} else {
			// Create destination file
			dest, err := os.Create(destinationPath)
			if err != nil {
				log.Fatalln("Unable to create destination:", destinationPath)
			}
			defer dest.Close()

			// Open input stream
			source, err := zipFile.Open()
			if err != nil {
				log.Fatalln("Unable to open input ZIP file:", zipFile.Name)
			}
			defer source.Close()

			// Extract file
			io.Copy(dest, source)
		}
	}
}

// Checks that all installer URLs are still reachable. Exits with an error on first failure.
func checkLinks(registry registry) {
	var errors []string

	checkLink := func(rawurl string) bool {
		response, err := http.Get(rawurl)
		if err != nil {
			return false
		}
		defer response.Body.Close()

		return response.StatusCode == http.StatusOK
	}

	checkArch := func(name string, version string, architecture string, rawUrl string) {
		if rawUrl == "" {
			return
		}

		url := strings.Replace(rawUrl, "${version}", version, -1)

		if !checkLink(url) {
			errors = append(errors, fmt.Sprintf("%v (%v): %v", name, architecture, url))
		}
	}

	for _, name := range sortedKeys(registry.Packages) {
		entry := registry.Packages[name]

		checkArch(name, entry.Version, "x86", entry.Installer.X86)
		checkArch(name, entry.Version, "x86_64", entry.Installer.X86_64)
	}

	if len(errors) > 0 {
		fmt.Println("\n\nThese installers were unreachable:")

		for _, e := range errors {
			fmt.Println(e)
		}

		os.Exit(1)
	}
}
