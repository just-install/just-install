package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-dry"
)

func main() {
	clean()
	build()
	buildMsi()

	if isStableBuild() {
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
	version, err := getVersion()
	if err != nil {
		log.Fatalln("could not get version string:", err)
	}

	log.Println("building version", version)

	cmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf("-s -w -X main.version=%s", version),
		"-trimpath", "./cmd/just-install")
	cmd.Env = append(os.Environ(), "GOARCH=386")
	if err := cmd.Run(); err != nil {
		log.Fatalln("cannot build just-install:", err)
	}
}

func buildMsi() {
	version, err := getVersion()
	if err != nil {
		log.Fatalln("could not get version string:", err)
	}

	log.Println("building MSI installer version", version)

	var env []string
	if isStableBuild() {
		env = append(os.Environ(), fmt.Sprintf("JUST_INSTALL_MSI_VERSION=%s", version))
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
	// FIXME: this thing is a mess, it's essentially a shell script within a Go program.
	log.Println("deploying")

	if err := run("git", "clone", "git@github.com:just-install/stable.git"); err != nil {
		log.Fatalf("could not clone stable repository: %v", err)
	}

	for _, f := range []string{"just-install.exe", "just-install.msi"} {
		if err := dry.FileCopy(f, fmt.Sprintf("stable\\%v", f)); err != nil {
			log.Fatalf("cannot copy %v to git repo: %v", f, err)
		}
	}

	if err := os.Chdir("stable"); err != nil {
		log.Fatalf("cannot chdir to git repo: %v", err)
	}

	if err := run("git", "add", "-A"); err != nil {
		log.Fatalf("cannot add deployment artifacts: %v", err)
	}

	if err := run("git", "commit", "--amend", "--no-edit", "--reset-author", "-m", "CI Release"); err != nil {
		log.Fatalf("unable to commit: %v", err)
	}

	if err := run("git", "push", "--force"); err != nil {
		log.Fatalf("cannot push to git repo: %v", err)
	}
}

func getVersion() (string, error) {
	out, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		return "", fmt.Errorf("could not get version string from git: %w", err)
	}

	str := string(out)

	if len(str) < 1 {
		return "", fmt.Errorf("invalid version string: %v", str)
	}

	if str[0] != 'v' {
		return "", fmt.Errorf("version string does not start with 'v': %v", str)
	}

	return strings.TrimSpace(str[1:]), nil
}

func isStableBuild() bool {
	ref, ok := dry.EnvironMap()["GITHUB_REF"]
	return ok && strings.HasPrefix(ref, "refs/tags/")
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(cmd.Env,
		"GIT_AUTHOR_EMAIL=CI",
		"GIT_AUTHOR_NAME=CI",
		"GIT_COMMITTER_EMAIL=CI",
		"GIT_COMMITTER_NAME=CI",
		// HACK: For some reason we can pull with a valid known_hosts file but pushing raises an
		// error, hence we disable host key checking for now. Since we are going from github.com to
		// github.com this is probably OK.
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no",
	)

	return cmd.Run()
}
