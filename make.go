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

	dry "github.com/ungerik/go-dry"
)

const accessControlHeader = `
/*
	Access-Control-Allow-Origin: https://just-install.github.io
`

func main() {
	clean()
	build()
	buildMsi()

	if shouldDeploy() {
		deploy()
	} else {
		log.Println("skipping deployment to netlify")
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

	createDeployArchive()
	uploadDeployArchive(fmt.Sprintf("https://api.netlify.com/api/v1/sites/just-install-%s.netlify.com/deploys", target))
}

func createDeployArchive() {
	f, err := os.Create("deploy.zip")
	if err != nil {
		log.Fatalln("cannot create deployment archive:", err)
	}
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
	if err != nil {
		log.Fatalf("cannot create zip writer for entry %v: %v\n", path, err)
	}

	src, err := os.Open(path)
	if err != nil {
		log.Fatalf("cannot open %v for read: %v\n", path, err)
	}
	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatalf("cannot copy %v to deployment archive: %v\n", path, err)
	}
}

func zipWriteString(w *zip.Writer, path string, s string) {
	dst, err := w.Create(path)
	if err != nil {
		log.Fatalf("cannot create zip writer for entry %v: %v\n", path, err)
	}

	if _, err := dst.Write([]byte(s)); err != nil {
		log.Fatalf("cannot write string as entry %v: %v\n", path, err)
	}
}

func uploadDeployArchive(url string) {
	f, err := os.Open("deploy.zip")
	if err != nil {
		log.Fatalln("cannot open deploy archive:", err)
	}
	defer f.Close()

	req, err := http.NewRequest("POST", url, f)
	if err != nil {
		log.Fatalln("cannot create request to Netlify's endpoint:", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dry.EnvironMap()["NETLIFY_DEPLOY_TOKEN"]))
	req.Header.Add("Content-Type", "application/zip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("request to Netlify upload endpoint failed:", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("expected 200 OK, instead got", resp.StatusCode)
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
