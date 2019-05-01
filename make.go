package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	dry "github.com/ungerik/go-dry"
)

func main() {
	clean()
	build()
	buildMsi()

	if shouldDeploy() {
		deploy()
	}
}

func clean() {
	toRemove := []string{"just-install"}

	globs := []string{"*.exe", "*.msi", "*.wixobj", "*.wixpdb"}
	for _, glob := range globs {
		found, err := filepath.Glob(glob)
		if err != nil {
			log.Fatalln("cannot glob files to remove:", err)
		}

		toRemove = append(toRemove, found...)
	}

	for _, path := range toRemove {
		log.Println("deleting", path)
		if err := os.RemoveAll(path); err != nil {
			log.Fatalf("cannot remove %s: %s\n", path, err)
		}
	}
}

func build() {
	log.Println("building version", getVersion())

	cmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf("-s -w -X main.version=%s", getVersion()),
		"./cmd/just-install")
	cmd.Env = append(os.Environ(), "GOARCH=386")
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot build just-install:", err)
	}
}

func buildMsi() {
	log.Println("building MSI installer")

	var env []string
	if isStableBuild() {
		env = append(os.Environ(), fmt.Sprintf("JUST_INSTALL_MSI_VERSION=%s", getVersion()))
	} else {
		env = append(os.Environ(), "JUST_INSTALL_MSI_VERSION=255.0")
	}

	cmd := exec.Command("candle", "just-install.wxs")
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot build MSI installer:", err)
	}

	cmd = exec.Command("light", "just-install.wixobj")
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot link MSI installer:", err)
	}
}

func deploy() {
	var target string
	if isStableBuild() {
		target = "stable"
	} else {
		target = "unstable"
	}

	log.Println("deploying to", target)

	ghPagesDeploy()
}

func ghPagesDeploy() {
	// TODO: this thing is a mess, it's essentially a shell script within a Go program.
	if !dry.FileExists("stable") {
		log.Fatalln("must clone git@github.com:just-install/stable.git")
	}

	for _, f := range []string{"just-install.exe", "just-install.msi"} {
		if err := dry.FileCopy(f, fmt.Sprintf("stable\\%v", f)); err != nil {
			log.Fatalln("cannot copy", f, "to git repo")
		}
	}

	if err := os.Chdir("stable"); err != nil {
		log.Fatalln("cannot chdir to git repo")
	}

	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot add deployment artifacts")
	}

	cmd = exec.Command("git", "commit", "--amend", "--no-edit", "--reset-author", "-m", "AppVeyor Release")
	if err := cmd.Run(); err != nil {
		log.Fatalln("unable to commit")
	}

	cmd = exec.Command("git", "push", "-f")
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot push to git repo")
	}
}

func getVersion() string {
	if !isStableBuild() {
		return "unstable"
	}

	f, err := dry.FileGetJSON(".releng.json")
	if err != nil {
		log.Fatalln("cannot read .releng.json:", err)
	}

	ret, ok := f.(map[string]interface{})["version"]
	if !ok {
		log.Fatalln("cannot read version from .releng.json")
	}

	return ret.(string)
}

func shouldDeploy() bool {
	// Skip pull requests
	_, ok := dry.EnvironMap()["APPVEYOR_PULL_REQUEST_NUMBER"]
	if ok {
		return false
	}

	// Only deploy stable builds, unstable builds will go to GitHub Releases.
	return isStableBuild()
}

func isStableBuild() bool {
	val, ok := dry.EnvironMap()["APPVEYOR_REPO_TAG_NAME"]
	return ok && len(val) > 0 && val != "unstable"
}
