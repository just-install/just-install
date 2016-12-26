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

	"github.com/ungerik/go-dry"
)

const (
	registrySupportedVersion = 4
	registryURL              = "http://registry.just-install.it"
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
	Interactive bool
	Kind        string
	Options     map[string]interface{} // Optional
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
	url := e.installerURL(arch)

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
func (e *RegistryEntry) JustInstall(force bool) {
	options := e.Installer.options()
	downloadedFile := e.DownloadInstaller(force)

	if container, ok := options["container"]; ok {
		tempDir := e.unwrapZip(downloadedFile) // Assuming it is a zip due to JSON schema
		installer := container.(map[string]interface{})["installer"].(string)

		e.install(filepath.Join(tempDir, installer))
	} else {
		e.install(downloadedFile)
	}

	e.CreateShims()
}

func (e *RegistryEntry) installerURL(arch string) string {
	var url string

	if arch == "x86_64" && e.Installer.X86_64 != "" {
		url = e.Installer.X86_64
	} else {
		url = e.Installer.X86
	}

	return e.ExpandString(url)
}

func (e *RegistryEntry) ExpandString(s string) string {
	return expandString(s, map[string]string{"version": e.Version})
}

func (e *RegistryEntry) unwrapZip(containerPath string) string {
	extractTo := filepath.Join(os.TempDir(), crc32s(containerPath))

	extractZip(containerPath, extractTo)

	return extractTo
}

func (e *RegistryEntry) install(installer string) {
	switch e.Installer.Kind {
	case "advancedinstaller":
		system(installer, "/q", "/i")
	case "as-is":
		system(installer)
	case "easy_install_26":
		system("\\Python26\\Scripts\\easy_install.exe", installer)
	case "easy_install_27":
		system("\\Python27\\Scripts\\easy_install.exe", installer)
	case "innosetup":
		system(installer, "/norestart", "/sp-", "/verysilent")
	case "msi":
		system("msiexec.exe", "/q", "/i", installer, "ALLUSERS=1", "REBOOT=ReallySuppress")
	case "nsis":
		system(installer, "/S", "/NCRC")
	case "custom":
		var args []string

		for _, v := range e.Installer.options()["arguments"].([]interface{}) {
			args = append(args, expandString(v.(string), map[string]string{"installer": installer}))
		}

		system(args...)
	case "copy":
		destination := e.destination()

		parentDir := filepath.Dir(destination)
		log.Println("Creating", parentDir)
		os.MkdirAll(parentDir, os.ModePerm)

		log.Println("Copying to", destination)
		dry.FileCopy(installer, destination)
	case "zip":
		destination := e.destination()

		log.Println("Extracting to", destination)

		extractZip(installer, destination)
	default:
		log.Fatalln("Unknown installer type:", e.Installer.Kind)
	}
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
		os.MkdirAll(shimsPath, 0)
	}

	if shims, ok := e.Installer.options()["shims"]; ok {
		for _, v := range shims.([]interface{}) {
			shimTarget := e.expandString(v.(string))
			shim := filepath.Join(shimsPath, filepath.Base(shimTarget))

			if dry.FileExists(shim) {
				os.Remove(shim)
			}

			log.Printf("Creating shim for %s (%s)\n", shimTarget, shim)

			system(exeproxy, "exeproxy-copy", shim, shimTarget)
		}
	}
}
