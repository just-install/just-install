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

package paths

import (
	"os"
	"path/filepath"
)

// TempFileCreate is the same as TempFile() but also creates just-install's temporary directory if
// missing.
func TempFileCreate(file string) (string, error) {
	if err := os.MkdirAll(TempDir(), 0700); err != nil {
		return "", err
	}

	return TempFile(file), nil
}

// TempFile returns the path to a temporary file below just-install's temporary file directory.
func TempFile(file string) string {
	return filepath.Join(TempDir(), file)
}

// TempDirCreate is the same as TempDir() but also creates just-install's temporary directory if
// missing.
func TempDirCreate() (string, error) {
	ret := TempDir()

	if err := os.MkdirAll(ret, 0700); err != nil {
		return "", err
	}

	return ret, nil
}

// TempDir returns the temporary directory that must be used to store all of just-install's files.
func TempDir() string {
	return filepath.Join(os.TempDir(), "just-install")
}
