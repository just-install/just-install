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
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/just-install/just-install/pkg/paths"
)

func handleCleanAction(c *cli.Context) {
	// Yup, this is weird, but we don't want a public API that allows us to use the temporary
	// directory before creating it elsewhere in the program.
	tempDir, err := paths.TempDirCreate()
	if err != nil {
		log.Fatalln("Could not create temporary directory:", err)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		log.Fatalln("Could not clean temporary directory:", err)
	}
}
