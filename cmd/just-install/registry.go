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

package main

import (
	"log"
	"os"
	"time"

	"github.com/just-install/just-install/pkg/fetch"
	"github.com/just-install/just-install/pkg/justinstall"
	"github.com/just-install/just-install/pkg/paths"
	dry "github.com/ungerik/go-dry"
	"github.com/urfave/cli"
)

const registryURL = "https://just-install.github.io/registry/just-install-v4.json"

func loadRegistry(c *cli.Context) justinstall.Registry {
	src := registryURL
	dst := paths.TempFile("registry.json")

	if c.GlobalIsSet("registry") {
		src = c.GlobalString("registry")
		dst = paths.TempFile("registry-custom.json")
	}

	if c.GlobalBool("force") && dry.FileExists(dst) {
		if err := os.Remove(dst); err != nil {
			log.Fatalln("Could not delete", dst, "due to", err)
		}
	}

	download := !dry.FileExists(dst)
	download = download || dry.FileTimeModified(dst).Before(time.Now().Add(-24*time.Hour))

	if !download {
		return justinstall.LoadRegistry(dst)
	}

	dst, err := fetch.Fetch(src, &fetch.Options{Destination: dst, Progress: true})
	if err != nil {
		log.Fatalln("Error obtaining registry:", err)
	}

	return justinstall.LoadRegistry(dst)
}
