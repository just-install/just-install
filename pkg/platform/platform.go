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

package platform

import (
	"os"
	"strings"

	"github.com/ungerik/go-dry"
)

// SetNormalisedProgramFilesEnv ensures that we have "%ProgramFiles%" and "%ProgramFiles(x86)"
// enviroment variables exported on both 32-bit and 64-bit Windows and pointing to something
// sensible on both architectures.
//
// This allows environment variable expansion for some strings in the registry to apply uniformly on
// both 32-bit and 64-bit Windows, even though we are a 32-bit process, since on 32-bit
// architectures "%ProgramFiles(x86)%" is missing.
//
// A call to this function should be made early in the program (possibily inside the main()
// function).
func SetNormalisedProgramFilesEnv() {
	var programFiles string
	var programFilesX86 string

	if Is64Bit() {
		// FIXME: We have to improvise here since we are a 32-bit process on 64-bit Windows. Find a
		// way to reliably get this information from Windows, without resorting to cgo.
		programFilesX86 = os.Getenv("ProgramFiles(x86)")
		programFiles = programFilesX86[0:strings.LastIndex(programFilesX86, " (x86)")]
	} else {
		programFiles = os.Getenv("ProgramFiles")
		programFilesX86 = programFiles
	}

	os.Setenv("ProgramFiles", programFiles)
	os.Setenv("ProgramFiles(x86)", programFilesX86)
}

// Is64Bit determines whether we are running on 32-bit or 64-bit Windows.
//
// This function performs platform identification by checking the presence of the
// "%ProgramFiles(x86)%" directory. This is because we compile a 32-bit binary intended to run on
// both 32-bit and 64-bit Windows and the usual detection mechanism may get confused when SysWOW64
// gets in the way. This is kind of an ugly hack.
func Is64Bit() bool {
	sentinel := os.Getenv("ProgramFiles(x86)")

	return len(sentinel) > 0 && dry.FileIsDir(sentinel)
}
