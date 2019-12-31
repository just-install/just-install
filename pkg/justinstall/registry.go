package justinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/just-install/just-install/pkg/cmd"
	"github.com/just-install/just-install/pkg/installer"
	dry "github.com/ungerik/go-dry"
)

const (
	registrySupportedVersion = 4
	registryURL              = "https://just-install.github.io/registry/just-install-v4.json"
)

var (
	arch         = "x86"
	isAmd64      = false
	shimsPath    = os.ExpandEnv("${SystemDrive}\\Shims")
	shimsPathOld = os.ExpandEnv("${SystemDrive}\\just-install")
	tempPath     = filepath.Join(os.TempDir(), "just-install")
	registryPath = filepath.Join(tempPath, fmt.Sprintf("just-install-v%v.json", registrySupportedVersion))
)

//
// Package Initialization
//

func init() {
	createTempDir()
	determineArch()
	normalizeProgramFiles()
}

func createTempDir() {
	os.MkdirAll(tempPath, 0700)
}

// determineArch determines the Windows architecture of the current Windows installation. It changes
// both the "isAmd64" and "arch" globals.
func determineArch() {
	// Since our output is a 32-bit executable (for maximum compatibility) and all other options
	// proved fruitless, let's just test for something that is usually available only on x86_64
	// editions of Windows.
	sentinel := os.Getenv("ProgramFiles(x86)")

	if sentinel == "" && !dry.FileIsDir(sentinel) {
		arch = "x86"
		isAmd64 = false
	} else {
		arch = "x86_64"
		isAmd64 = true
	}
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

//
// Public
//

func CleanTempDir() error {
	if err := os.RemoveAll(tempPath); err != nil {
		return err
	}

	createTempDir()

	return nil
}

// SetArchitecture changes the architecture of future package installations.
func SetArchitecture(a string) error {
	if a == "x86_64" && !isAmd64 {
		return errors.New("This machine is not 64-bit capable")
	} else if a != "x86" && a != "x86_64" {
		return fmt.Errorf("Unknown architecture: %v", a)
	}

	arch = a

	return nil
}

// SmartLoadRegistry tries to load a cached copy downloaded from the Internet. If neither is
// available, it tries to download it from the known location first.
func SmartLoadRegistry(force bool) Registry {
	download := !dry.FileExists(registryPath)
	download = download || dry.FileTimeModified(registryPath).Before(time.Now().Add(-24*time.Hour))
	download = download || force

	if download {
		log.Println("Updating registry from:", registryURL)

		downloadRegistry()
	}

	return LoadRegistry(registryPath)
}

// LoadRegistry unmarshals the registry from a local file path.
func LoadRegistry(path string) Registry {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read the registry file.")
	}

	var ret Registry

	if err := json.Unmarshal(data, &ret); err != nil {
		log.Fatalln("Unable to parse the registry file.")
	}

	if ret.Version != registrySupportedVersion {
		log.Fatalln("Please update to a new version of just-install by running: msiexec.exe /i https://just-install.github.io/stable/just-install.msi")
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
	Interactive bool
	Kind        string
	Options     map[string]interface{} // Optional
	Preinstall  []string // Optional
	Postinstall []string // Optional
	X86         string
	X86_64      string
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

// Registry is a list of packages that just-install knows how to install.
type Registry struct {
	Version  int
	Packages map[string]RegistryEntry
}

// SortedPackageNames returns the list of packages present in the registry, sorted alphabetically.
func (r *Registry) SortedPackageNames() []string {
	var keys []string

	for k := range r.Packages {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// RegistryEntry is a single entry in the just-install registry.
type RegistryEntry struct {
	Version   string
	Installer installerEntry
}

// DownloadInstaller downloads the installer for the current entry in the temporary directory.
func (e *RegistryEntry) DownloadInstaller(force bool) string {
	options := e.Installer.options()

	url, err := e.installerURL(arch)
	if err != nil {
		log.Fatalln("Cannot download installation package:", err)
	}

	log.Println(arch, "-", url)

	if filename, ok := options["filename"]; ok {
		return downloadTemp(url, filename.(string), force)
	} else if ext, ok := options["extension"]; ok {
		return downloadExt(url, ext.(string), force)
	}

	return downloadAutoExt(url, force)
}

// JustInstall will download and install the given registry entry. Setting `force` to true will
// force a re-download and re-installation the package.
func (e *RegistryEntry) JustInstall(force bool) error {
	options := e.Installer.options()
	downloadedFile := e.DownloadInstaller(force)

	for _, command := range e.Installer.Preinstall {
		cmd.Run(strings.Fields(command)...)
	}

	if container, ok := options["container"]; ok {
		tempDir := filepath.Join(os.TempDir(), crc32s(downloadedFile))
		if err := installer.ExtractZIP(downloadedFile, tempDir); err != nil {
			return err
		}

		installer := container.(map[string]interface{})["installer"].(string)
		if err := e.install(filepath.Join(tempDir, installer)); err != nil {
			return err
		}
	} else {
		if err := e.install(downloadedFile); err != nil {
			return err
		}
	}

	for _, command := range e.Installer.Postinstall {
		cmd.Run(strings.Fields(command)...)
	}

	e.CreateShims()

	return nil
}

func (e *RegistryEntry) installerURL(arch string) (string, error) {
	var url string

	if arch == "x86_64" {
		if e.Installer.X86_64 != "" {
			url = e.Installer.X86_64
		} else if e.Installer.X86 != "" {
			url = e.Installer.X86
		} else {
			return "", errors.New("No fallback 32-bit download")
		}
	} else if arch == "x86" {
		if e.Installer.X86 != "" {
			url = e.Installer.X86
		} else {
			return "", errors.New("64-bit only package")
		}
	} else {
		return "", errors.New("Unknown architecture")
	}

	return e.ExpandString(url), nil
}

func (e *RegistryEntry) ExpandString(s string) string {
	return expandString(s, map[string]string{"version": e.Version})
}

func (e *RegistryEntry) install(path string) error {
	if e.Installer.Kind == "custom" {
		var args []string

		for _, v := range e.Installer.options()["arguments"].([]interface{}) {
			args = append(args, expandString(v.(string), map[string]string{"installer": path}))
		}

		return cmd.Run(args...)
	}

	installerType := installer.InstallerType(e.Installer.Kind)
	if !installerType.IsValid() {
		return fmt.Errorf("unknown installer type: %v", e.Installer.Kind)
	}

	return cmd.Run(installer.Command(path, installerType)...)
}

func (e *RegistryEntry) destination() string {
	return expandString(os.ExpandEnv(e.Installer.options()["destination"].(string)), nil)
}

func (e *RegistryEntry) CreateShims() {
	exeproxy := os.ExpandEnv("${ProgramFiles(x86)}\\exeproxy\\exeproxy.exe")
	if !dry.FileExists(exeproxy) {
		return
	}

	if dry.FileIsDir(shimsPathOld) {
		fmt.Println("")
		fmt.Println("**************************************************************************")
		fmt.Println("Shims are now placed under " + shimsPath)
		fmt.Println("")
		fmt.Println("We have left your old shims at " + shimsPathOld + " but we are creating")
		fmt.Println("the new ones at the aforementioned path and we invite you to move them")
		fmt.Println("and update your %PATH% variable to include the new path.")
		fmt.Println("**************************************************************************")
		fmt.Println("")
	}

	if !dry.FileIsDir(shimsPath) {
		if err := os.MkdirAll(shimsPath, 0); err != nil {
			// FIXME: add proper error handling
			log.Fatalln("Could not create shim directory:", err)
		}
	}

	if shims, ok := e.Installer.options()["shims"]; ok {
		for _, v := range shims.([]interface{}) {
			shimTarget := e.ExpandString(v.(string))
			shim := filepath.Join(shimsPath, filepath.Base(shimTarget))

			if dry.FileExists(shim) {
				os.Remove(shim)
			}

			log.Printf("Creating shim for %s (%s)\n", shimTarget, shim)

			if err := cmd.Run(exeproxy, "exeproxy-copy", shim, shimTarget); err != nil {
				// FIXME: add proper error handling
				log.Fatalln("Could not create shim:", err)
			}
		}
	}
}
