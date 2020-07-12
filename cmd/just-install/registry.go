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
	"fmt"
	"os"
	"time"

	"github.com/ungerik/go-dry"
	"github.com/urfave/cli/v2"

	"github.com/just-install/just-install/pkg/fetch"
	"github.com/just-install/just-install/pkg/justinstall"
	"github.com/just-install/just-install/pkg/paths"
)

const registryURL = "https://just-install.github.io/registry/just-install-v4.json"

func loadRegistry(c *cli.Context, force bool, progress bool) (*justinstall.Registry, error) {
	src := registryURL
	dst, dstErr := paths.TempFileCreate("registry.json")
	if dstErr != nil {
		return nil, fmt.Errorf("could not create temporary directory to hold registry file: %w", dstErr)
	}

	if c.IsSet("registry") {
		src = c.String("registry")
		dst, dstErr = paths.TempFileCreate("registry-custom.json")
		if dstErr != nil {
			return nil, fmt.Errorf("could not create temporary directory to hold custom registry file: %w", dstErr)
		}
	}

	if force && dry.FileExists(dst) {
		if err := os.Remove(dst); err != nil {
			return nil, fmt.Errorf("could not delete %v due to %w", dst, err)
		}
	}

	download := !dry.FileExists(dst)
	download = download || dry.FileTimeModified(dst).Before(time.Now().Add(-24*time.Hour))
	if !download {
		ret := justinstall.LoadRegistry(dst)
		return &ret, nil
	}

	dst, err := fetch.Fetch(src, &fetch.Options{Destination: dst, Progress: progress})
	if err != nil {
		return nil, fmt.Errorf("error obtaining registry: %w", err)
	}

	ret := justinstall.LoadRegistry(dst)
	return &ret, nil
}
