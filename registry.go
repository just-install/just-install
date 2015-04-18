package justinstall

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

	"github.com/cheggaaa/pb"
	"github.com/ungerik/go-dry"
)

const (
	registrySupportedVersion = 3
	registryURL              = "https://raw.github.com/lvillani/just-install/v3.0/just-install.json"
)

var (
	arch         = "x86"
	isAmd64      = false
	shimsPath    = os.ExpandEnv("${SystemDrive}\\just-install")
	registryPath = filepath.Join(os.TempDir(), "just-install.json")
)

func init() {
	determineArch()
	normalizeProgramFiles()
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

func (r *registry) SortedPackageNames() []string {
	var keys []string

	for k := range r.Packages {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
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

	for k := range m {
		keys = append(keys, k)
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
	defer progressBar.Finish()

	progressBar.ShowSpeed = true
	progressBar.SetRefreshRate(time.Millisecond * 1000)
	progressBar.SetUnits(pb.U_BYTES)
	progressBar.Start()

	writer := io.MultiWriter(destination, progressBar)

	io.Copy(writer, response.Body)
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
