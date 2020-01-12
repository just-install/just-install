// just-install - The simple package installer for Windows
// Copyright (C) 2019 just-install authors.
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

package fetch

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/ungerik/go-dry"
)

// HTTPStatusError describes an unexpected HTTP response status code.
type HTTPStatusError struct {
	Expected int
	Received int
	Resource string
}

func (h *HTTPStatusError) Error() string {
	return fmt.Sprintf("expected status code %v but received %v instead (%v)", h.Expected, h.Received, h.Resource)
}

// ContentTypeError describes an unexpected value of the Content-Type header.
type ContentTypeError struct {
	Received string
	Resource string
}

func (c *ContentTypeError) Error() string {
	return fmt.Sprintf("unexpected Content-Type %v (%v)", c.Received, c.Resource)
}

// Options that influence Fetch.
type Options struct {
	Destination string      // Can either be a file path or a directory path. If it's a directory, it must already exist.
	Progress    bool        // Whether to show the progress indicator.
	HTTP        HTTPOptions // HTTP client options.
}

// HTTPOptions contains cookies and headers to send when making an HTTP request.
type HTTPOptions struct {
	CheckRedirect func(req *http.Request, via []*http.Request) error // Same as http.Client.CheckRedirect
	Cookies       map[string][2]string                               // URL -> (Cookie name, Cookie Value)
	Headers       map[string]string                                  // Header -> Value
}

// CheckOptions are options that influence Check.
type CheckOptions struct {
	Options
	ExpectedContentTypes []string // Acceptable values for the Content-Type header
}

// Check returns true if running Fetch with the same resource has a high-chance of actually fetching
// it. This is mostly used by `just-install audit` to check whether the registry contains broken
// entries.
func Check(resource string, options *CheckOptions) error {
	// Shortcut: resource is a local file and we can return immediately
	if dry.FileExists(resource) {
		return nil
	}

	// Parse resource URL
	parsedURL, err := url.Parse(resource)
	if err != nil {
		return err
	}

	if parsedURL.Scheme == "file" {
		return errors.New("cannot check local files")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %v", parsedURL.Scheme)
	}

	// Options
	if options == nil {
		options = &CheckOptions{}
	}

	// Request
	resp, err := get(resource, &options.Options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &HTTPStatusError{http.StatusOK, resp.StatusCode, resource}
	}

	contentType := resp.Header.Get("Content-Type")
	if len(options.ExpectedContentTypes) > 0 && !dry.StringInSlice(contentType, options.ExpectedContentTypes) {
		return &ContentTypeError{contentType, resource}
	}

	return nil
}

// Fetch obtains the given resource, either a local file or something that can be download via
// HTTP/HTTPS, to a file on disk. Returns the path to the fetched file or an error.
func Fetch(resource string, options *Options) (string, error) {
	// Shortcut: resource is a local file and we can return its path immediately.
	if dry.FileExists(resource) {
		return resource, nil
	}

	// Parse resource URL
	parsedURL, err := url.Parse(resource)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme == "file" {
		return parsedURL.Path, nil
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %v", parsedURL.Scheme)
	}

	// Options
	if options == nil {
		options = &Options{}
	}

	if options.Destination == "" {
		return "", errors.New("destination must be either a file or directory path")
	}

	// Request
	var lastLocation *url.URL
	options.HTTP.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// This is the same check used by the CheckRedirect function used in the standard library
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}

		// Store the last redirect
		lastLocation = req.URL
		return nil
	}

	resp, err := get(resource, options)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &HTTPStatusError{http.StatusOK, resp.StatusCode, resource}
	}

	// Compute final destination path
	dest := options.Destination
	if dry.FileIsDir(dest) {
		contentDisposition := resp.Header.Get("Content-Disposition")
		re := regexp.MustCompile(`filename="?([\w\.\-]+)"?`)

		if re.MatchString(contentDisposition) {
			dest = filepath.Join(dest, re.FindStringSubmatch(contentDisposition)[1])
		} else if lastLocation == nil {
			dest = filepath.Join(dest, filepath.Base(parsedURL.Path))
		} else {
			dest = filepath.Join(dest, filepath.Base(lastLocation.Path))
		}
	}

	// File already exists, return its path.
	if dry.FileExists(dest) {
		return dest, nil
	}

	// Fetch to temporary file
	destTmp := dest + ".download"

	destTmpWriter, err := os.Create(destTmp)
	if err != nil {
		return "", err
	}
	defer destTmpWriter.Close()

	var copyWriter io.WriteCloser = destTmpWriter
	if options.Progress {
		log.Println("Fetching", resource, "to", dest)

		progressBar := pb.New64(resp.ContentLength)
		progressBar.Set(pb.Bytes, true)
		progressBar.SetRefreshRate(time.Second)
		defer progressBar.Finish()

		copyWriter = progressBar.NewProxyWriter(destTmpWriter)
		defer copyWriter.Close()

		progressBar.Start()
	}

	//==============================================================================================
	// NOTE: destTmpWriter and copyWriter may actually be the same thing here if `Options.Progress`
	// is false.
	//==============================================================================================

	if _, err := io.Copy(copyWriter, resp.Body); err != nil {
		return "", err
	}

	// Must explicitly close these before renaming the file, since defers run too late
	copyWriter.Close()
	destTmpWriter.Close()
	resp.Body.Close()

	// Move temporary file back to definitive place
	if err := os.Rename(destTmp, dest); err != nil {
		return "", err
	}

	return dest, nil
}

// get performs an HTTP GET request using our custom client and options.
func get(resource string, options *Options) (*http.Response, error) {
	req, err := http.NewRequest("GET", resource, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range options.HTTP.Headers {
		req.Header.Set(k, v)
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	for cookieURL, cookie := range options.HTTP.Cookies {
		u, err := url.Parse(cookieURL)
		if err != nil {
			return nil, fmt.Errorf("could not parse cookie URL: %v", cookieURL)
		}

		cookieJar.SetCookies(u, []*http.Cookie{&http.Cookie{Name: cookie[0], Value: cookie[1]}})
	}

	httpClient := NewClient()
	httpClient.CheckRedirect = options.HTTP.CheckRedirect
	httpClient.Jar = cookieJar

	return httpClient.Do(req)
}
