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

package installer

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ExtractZIP extracts the given ZIP archive to the given destination directory. If the destination
// directory does not exist, it is created.
func ExtractZIP(path string, dest string) error {
	if err := os.MkdirAll(dest, 0700); err != nil {
		return err
	}

	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		destinationPath := filepath.Join(dest, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			if err := os.MkdirAll(destinationPath, zipFile.Mode()); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(destinationPath), 0700); err != nil {
				return err
			}

			dest, err := os.Create(destinationPath)
			if err != nil {
				return err
			}
			defer dest.Close()

			source, err := zipFile.Open()
			if err != nil {
				return err
			}
			defer source.Close()

			if _, err := io.Copy(dest, source); err != nil {
				return err
			}

			// Must explicitly close dest and source at each iteration, defers run too late.
			dest.Close()
			source.Close()
		}
	}

	return nil
}
