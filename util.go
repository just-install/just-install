package justinstall

import (
	"archive/zip"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/ungerik/go-dry"
)

// expandString expands any environment variable in the given string, with additonal variables
// coming from the given context.
func expandString(s string, context map[string]string) string {
	data := environMap()

	// Merge the given context
	for k, v := range context {
		data[k] = v
	}

	var buf bytes.Buffer

	template.Must(template.New("expand").Parse(s)).Execute(&buf, data)

	return buf.String()
}

// environMap returns the current environment variables as a map.
func environMap() map[string]string {
	ret := make(map[string]string)
	env := os.Environ()

	for _, v := range env {
		split := strings.SplitN(v, "=", 2)

		if split[0] == "" && split[1] == "" {
			continue
		}

		split[0] = strings.ToUpper(split[0]) // Normalize variable names to upper case
		split[0] = strings.Replace(split[0], "(X86)", "_X86", -1)

		ret[split[0]] = split[1]
	}

	return ret
}

func system(args ...string) {
	var cmd *exec.Cmd

	if len(args) == 0 {
		return
	} else if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	log.Println("Running", strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// Convenience wrapper over download3 which passes an empty ("") `ext` parameter.
func downloadAutoExt(rawurl string, force bool) string {
	return downloadExt(rawurl, "", force)
}

// Downloads a file over HTTP(S) to a temporary location. The temporary file has a name derived
// from the CRC32 of the URL string with the original file extension attached (if any). If `ext`
// is not the empty string, it will be appended to the destination file. The file is re-downloaded
// only if the temporary file is missing or `force` is true.
func downloadExt(rawurl string, ext string, force bool) string {
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

	return downloadTemp(rawurl, base, force)
}

// Computes and returns the CRC32 of a string as an HEX string.
func crc32s(s string) string {
	crc32 := crc32.NewIEEE()
	crc32.Write([]byte(s))

	return fmt.Sprintf("%X", crc32.Sum32())
}

// downloadTemp downloads a file to the machine's temporary directory.
func downloadTemp(rawurl string, filename string, force bool) string {
	ret := filepath.Join(tempPath, filename)

	maybeDownload(rawurl, ret, force)

	return ret
}

// maybeDownload is a wrapper for download that doesn't re-download an existing file unless
// forced.
func maybeDownload(rawurl string, destinationPath string, force bool) {
	if !dry.FileExists(destinationPath) || force {
		download(rawurl, destinationPath)
	}
}

// download a file with the HTTP/HTTPS protocol showing a progress bar. The destination file is
// always overwritten.
func download(rawurl string, destinationPath string) {
	tempDestinationPath := destinationPath + ".tmp"

	destination, err := os.Create(tempDestinationPath)
	if err != nil {
		log.Fatalf("Unable to open the destination file: %s", tempDestinationPath)
	}
	defer destination.Close()

	response, err := CustomGet(rawurl)
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
	destination.Close()
	os.Rename(tempDestinationPath, destinationPath)
}

func CustomGet(urlStr string) (*http.Response, error) {
	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Codeplex
	if strings.Contains(urlStr, "download-codeplex.sec.s-msft.com") {
		request.Header.Set("User-Agent", "chocolatey command line")
	}

	// AMD Catalyst
	if strings.Contains(urlStr, "ati.com") {
		request.Header.Set("Referer", "http://support.amd.com/")
	}

	// JRE/JDK from java.oracle.com
	oracleURL, _ := url.Parse("http://download.oracle.com")
	oracleEdeliveryURL, _ := url.Parse("https://edelivery.oracle.com")
	oracleCookies := []*http.Cookie{{Name: "oraclelicense", Value: "accept-securebackup-cookie"}}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(oracleURL, oracleCookies)
	jar.SetCookies(oracleEdeliveryURL, oracleCookies)

	client := http.Client{Jar: jar}

	return client.Do(request)
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
