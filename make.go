package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ungerik/go-dry"
)

const accessControlHeader = `
/*
	Access-Control-Allow-Origin: https://just-install.it
`

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
		checkErr(err, "cannot glob files to remove")
		toRemove = append(toRemove, found...)
	}

	for _, path := range toRemove {
		log.Println("deleting", path)
		checkErr(os.RemoveAll(path), "cannot remove %s", path)
	}
}

func build() {
	log.Println("building version", getVersion())
	cmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf("-s -w -X main.version=%s", getVersion()),
		"-mod=readonly",
		"./cmd/just-install")
	cmd.Env = append(os.Environ(), "GOARCH=386")
	checkErr(cmd.Run(), "cannot build just-install")
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
	checkErr(cmd.Run(), "cannot build MSI installer")

	cmd = exec.Command("light", "just-install.wixobj")
	cmd.Env = env
	checkErr(cmd.Run(), "cannot link MSI installer")
}

func deploy() {
	var target string
	if isStableBuild() {
		target = "stable"
	} else {
		target = "unstable"
	}

	log.Println("deploying to", target)

	createDeployArchive()
	uploadDeployArchive(fmt.Sprintf("https://api.netlify.com/api/v1/sites/just-install-%s.netlify.com/deploys", target))
}

func createDeployArchive() {
	f, err := os.Create("deploy.zip")
	checkErr(err, "cannot create deployment archive")
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, path := range []string{"just-install.exe", "just-install.msi"} {
		zipCopy(w, path)
	}

	zipWriteString(w, "_redirects", "/    /just-install.msi    302")
	zipWriteString(w, "_headers", accessControlHeader)
}

func zipCopy(w *zip.Writer, path string) {
	dst, err := w.Create(path)
	checkErr(err, "cannot create zip writer for entry %s", path)

	src, err := os.Open(path)
	checkErr(err, "cannot open %s for reading", path)
	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		checkErr(err, "cannot copy %s to deployment archive", path)
	}
}

func zipWriteString(w *zip.Writer, path string, s string) {
	dst, err := w.Create(path)
	checkErr(err, "cannot create zip writer for entry %s", path)

	if _, err := dst.Write([]byte(s)); err != nil {
		checkErr(err, "cannot write string as entry %s", path)
	}
}

func uploadDeployArchive(url string) {
	f, err := os.Open("deploy.zip")
	checkErr(err, "cannot open deploy archive")
	defer f.Close()

	req, err := http.NewRequest("POST", url, f)
	checkErr(err, "cannot create request to netlify")

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dry.EnvironMap()["NETLIFY_DEPLOY_TOKEN"]))
	req.Header.Add("Content-Type", "application/zip")

	resp, err := http.DefaultClient.Do(req)
	checkErr(err, "upload to netlify failed")

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("expected 200 OK, instead got", resp.StatusCode)
	}
}

func getVersion() string {
	if !isStableBuild() {
		return "unstable"
	}

	f, err := dry.FileGetJSON(".releng.json")
	checkErr(err, "cannot read .releng.json")

	ret, ok := f.(map[string]interface{})["version"]
	if !ok {
		log.Fatalln("cannot read version from .releng.json")
	}

	return ret.(string)
}

func shouldDeploy() bool {
	_, ok := dry.EnvironMap()["APPVEYOR_PULL_REQUEST_NUMBER"]
	return !ok
}

func isStableBuild() bool {
	val, ok := dry.EnvironMap()["APPVEYOR_REPO_TAG_NAME"]
	return ok && len(val) > 0
}

func checkErr(err error, message string, v ...interface{}) {
	if err != nil {
		log.Fatalf("%s: %v\n", fmt.Sprintf(message, v...), err)
	}
}
