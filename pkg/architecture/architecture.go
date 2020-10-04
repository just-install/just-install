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

package architecture

// TODO: Turn this into an enum

// Supported architectures.
const (
	X86    = "x86"
	X86_64 = "x86_64"
)

// IsValid returns whether the given string can be converted to a valid Architecture.
func IsValid(s string) bool {
	switch s {
	case X86, X86_64:
		return true
	default:
		return false
	}
}

// Architectures returns all the supported architectures.
func Architectures() []string {
	return []string{X86, X86_64}
}
