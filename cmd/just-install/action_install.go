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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gotopkg/mslnk/pkg/mslnk"
	"github.com/ungerik/go-dry"
	"github.com/urfave/cli/v2"

	"github.com/just-install/just-install/pkg/cmd"
	"github.com/just-install/just-install/pkg/fetch"
	"github.com/just-install/just-install/pkg/installer"
	"github.com/just-install/just-install/pkg/paths"
	"github.com/just-install/just-install/pkg/platform"
	"github.com/just-install/just-install/pkg/registry4"
	"github.com/just-install/just-install/pkg/strings2"
)

var (
	shimsPath = os.ExpandEnv("${SystemDrive}\\Shims")
	startMenu = os.ExpandEnv("${ProgramData}\\Microsoft\\Windows\\Start Menu\\Programs")
)

func handleInstall(c *cli.Context) error {
	ignoreCache := c.Bool("ignore-cache")
	onlyDownload := c.Bool("download-only")
	onlyShims := c.Bool("shim")
	progress := !c.Bool("noprogress")

	// We are explicitly NOT using the value of ignoreCache here (thus passing "false"
	// to loadRegistry's "force" argument), since forcing a registry download is done
	// via the "update" action.
	registry, err := loadRegistry(c, false, progress)
	if err != nil {
		return err
	}

	arch, err := getInstallArch(c.String("arch"))
	if err != nil {
		return err
	}

	lang := c.String("lang")
	if lang == "" {
		lang = "en-US"
	}

	// Install packages
	hasErrors := false

	for _, pkg := range c.Args().Slice() {
		entry, ok := registry.Packages[pkg]
		if !ok {
			log.Println("WARNING: unknown package", pkg)
			continue
		}

		options, err := entry.Installer.OptionsForArch(arch)
		if err != nil {
			return err
		}

		if onlyShims {
			if err := createShims(options); err != nil {
				return err
			}

			continue
		}

		installerPath, err := fetchInstaller(entry, arch, lang, ignoreCache, progress)
		if err != nil {
			log.Printf("error downloading %v: %v", pkg, err)
			hasErrors = true
			continue
		}

		if onlyDownload {
			continue
		}

		installerPath, err = maybeExtractContainer(installerPath, options)
		if err != nil {
			return err
		}

		if err := install(installerPath, entry.Installer.Kind, options); err != nil {
			log.Printf("error installing %v: %v", pkg, err)
			hasErrors = true
			continue
		}

		if exeproxyExists() {
			createShims(options)
		}
	}

	if hasErrors {
		return errors.New("encountered errors installing packages (see the log for details)")
	}

	return nil
}

// getInstallArch returns the architecture selected for package installation based on the given
// preferred architecture (e.g. given by the user via command line arguments). The given preferred
// architecture can be empty, in which case a suitable one is automatically selected for the current
// machine.
func getInstallArch(preferredArch string) (string, error) {
	switch preferredArch {
	case "":
		if platform.Is64Bit() {
			return "x86_64", nil
		}

		return "x86", nil
	case "x86":
		return preferredArch, nil
	case "x86_64":
		if !platform.Is64Bit() {
			return "", errors.New("this machine cannot run 64-bit software")
		}

		return preferredArch, nil
	default:
		return "", fmt.Errorf("unknown architecture: %v", preferredArch)
	}
}

// fetchInstaller fetches the installer for the given package and returns
func fetchInstaller(entry *registry4.Package, arch string, lang string, overwrite bool, progress bool) (string, error) {
	// Sanity check
	if isEmptyString(entry.Installer.X86) && isEmptyString(entry.Installer.X86_64) {
		return "", errors.New("package entry is missing both 32-bit and 64-bit installers")
	}

	// Pick preferred installer
	var installerURL string
	switch arch {
	case "x86":
		if isEmptyString(entry.Installer.X86) {
			return "", errors.New("this package doesn't offer a 32-bit installer")
		}

		installerURL = entry.Installer.X86
	case "x86_64":
		if isEmptyString(entry.Installer.X86_64) {
			// Fallback to the 32-bit installer
			installerURL = entry.Installer.X86
		} else {
			installerURL = entry.Installer.X86_64
		}
	default:
		panic("programmer error")
	}

	installerURL, err := expandString(installerURL, map[string]string{"version": entry.Version, "lang": lang})
	if err != nil {
		return "", fmt.Errorf("could not expand installer URL's template string: %w", err)
	}

	downloadDir, err := paths.TempDirCreate()
	if err != nil {
		return "", fmt.Errorf("could not create temporary directory to download installer: %w", err)
	}

	ret, err := fetch.Fetch(installerURL, &fetch.Options{
		Destination: downloadDir,
		Overwrite:   overwrite,
		Progress:    progress,
	})

	return ret, err
}

func maybeExtractContainer(path string, options *registry4.Options) (string, error) {
	if options == nil || options.Container == nil {
		return path, nil
	}

	if options.Container.Kind != "zip" {
		return "", errors.New("only \"zip\" containers are supported")
	}

	tempDir, err := paths.TempDirCreate()
	if err != nil {
		return "", err
	}

	extractDir := filepath.Join(tempDir, filepath.Base(path)+"_extracted")
	log.Println("extracting container", path, "to", extractDir)
	if err := installer.ExtractZIP(path, extractDir); err != nil {
		return "", err
	}

	if strings2.IsEmpty(options.Container.Installer) {
		files, err := ioutil.ReadDir(extractDir)
		if err != nil {
			return "", err
		}

		if len(files) == 1 {
			return filepath.Join(extractDir, files[0].Name()), nil
		} else {
			return "", errors.New("\"installer\" option is empty and container contains more than one file")
		}
	} else {
		return filepath.Join(extractDir, options.Container.Installer), nil
	}
}

func install(path string, kind string, options *registry4.Options) error {
	// One-off, custom, installers
	switch kind {
	case "copy":
		if options == nil {
			return errors.New("the \"copy\" installer requires additional options")
		}

		if strings2.IsEmpty(options.Destination) {
			return errors.New("\"destination\" is missing from installer options")
		}

		destination, err := expandString(options.Destination, nil)
		if err != nil {
			return fmt.Errorf("could not expand destination string: %w", err)
		}

		parentDir := filepath.Dir(destination)
		log.Println("creating", parentDir)
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			return err
		}

		log.Println("copying to", destination)
		return dry.FileCopy(path, destination)
	case "custom":
		if options == nil {
			return errors.New("the \"custom\" installer requires additional options")
		}

		if len(options.Arguments) < 1 {
			return errors.New("\"arguments\" is missing from installer options")
		}

		var args []string
		for _, v := range options.Arguments {
			expanded, err := expandString(v, map[string]string{"installer": path})
			if err != nil {
				return err
			}

			args = append(args, expanded)
		}

		return cmd.Run(args...)
	case "zip":
		if options == nil {
			return errors.New("the \"zip\" installer requires additional options")
		}

		if strings2.IsEmpty(options.Destination) {
			return errors.New("\"destination\" is missing from installer options")
		}

		destination, err := expandString(options.Destination, nil)
		if err != nil {
			return fmt.Errorf("could not expand destination string: %w", err)
		}

		log.Println("extracting to", destination)
		if err := installer.ExtractZIP(path, destination); err != nil {
			return err
		}

		for _, shortcut := range options.Shortcuts {
			shortcutName, err := expandString(shortcut.Name, nil)
			if err != nil {
				return fmt.Errorf("could not expand shortcut name string template: %w", err)
			}

			shortcutTarget, err := expandString(shortcut.Target, nil)
			if err != nil {
				return fmt.Errorf("could not expand shortcut target string template: %w", err)
			}

			shortcutLocation := filepath.Join(startMenu, shortcutName+".lnk")

			log.Println("creating shortcut to", shortcutTarget, "in", shortcutLocation)
			if err := mslnk.LinkFile(shortcutTarget, shortcutLocation); err != nil {
				return fmt.Errorf("could not create shortcut: %w", err)
			}
		}

		return nil
	}

	// Regular installer
	installerType := installer.InstallerType(kind)
	if !installerType.IsValid() {
		return fmt.Errorf("unknown installer type: %v", kind)
	}

	installerCommand, err := installer.Command(path, installerType)
	if err != nil {
		return err
	}

	return cmd.Run(installerCommand...)
}

func exeproxyExists() bool {
	exeproxy := os.ExpandEnv("${ProgramFiles(x86)}\\exeproxy\\exeproxy.exe")

	return dry.FileExists(exeproxy)
}

func createShims(options *registry4.Options) error {
	exeproxy := os.ExpandEnv("${ProgramFiles(x86)}\\exeproxy\\exeproxy.exe")

	if !dry.FileIsDir(shimsPath) {
		if err := os.MkdirAll(shimsPath, 0); err != nil {
			return fmt.Errorf("could not create shims directory %s: %w", shimsPath, err)
		}
	}

	for _, v := range options.Shims {
		shimTarget, err := expandString(v, map[string]string{})
		if err != nil {
			return err
		}

		shim := filepath.Join(shimsPath, filepath.Base(shimTarget))

		if dry.FileExists(shim) {
			if err := os.Remove(shim); err != nil {
				return fmt.Errorf("could not re-create shim %s: %w", shim, err)
			}
		}

		log.Printf("creating shim for %s (%s)\n", shimTarget, shim)

		if err := cmd.Run(exeproxy, "exeproxy-copy", shim, shimTarget); err != nil {
			return fmt.Errorf("could not create shim %s for %s: %w", shim, shimTarget, err)
		}
	}

	return nil
}

func isEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) < 1
}

// expandString expands any environment variable in the given string, with additional variables
// coming from the given context.
func expandString(s string, context map[string]string) (string, error) {
	data := environMap()

	// Merge the given context
	for k, v := range context {
		data[k] = v
	}

	var buf bytes.Buffer
	t, err := template.New("expandString").Parse(s)
	if err != nil {
		return "", err
	}
	t.Execute(&buf, data)

	return buf.String(), nil
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
