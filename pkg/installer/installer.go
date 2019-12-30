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

package installer

// InstallerType is a recognized installer type
type InstallerType string

// IsValid returns whether the given installer type is known.
func (it InstallerType) IsValid() bool {
	switch it {
	case AdvancedInstaller, AsIs, InnoSetup, MSI, NSIS, Squirrel:
		return true
	default:
		return false
	}
}

const (
	AdvancedInstaller InstallerType = "advancedinstaller"
	AsIs              InstallerType = "as-is"
	InnoSetup         InstallerType = "innosetup"
	MSI               InstallerType = "msi"
	NSIS              InstallerType = "nsis"
	Squirrel          InstallerType = "squirrel"
)

// Command returns the command needed to run the given installer of the given type.
func Command(path string, installerType InstallerType) []string {
	switch installerType {
	case AdvancedInstaller:
		return []string{path, "/i", "/q"}
	case AsIs:
		return []string{path}
	case InnoSetup:
		return []string{path, "/norestart", "/sp-", "/verysilent"}
	case MSI:
		return []string{"msiexec.exe", "/q", "/i", path, "ALLUSERS=1", "REBOOT=ReallySuppress"}
	case NSIS:
		return []string{path, "/S"}
	case Squirrel:
		return []string{path, "--silent"}
	default:
		panic("unknown installer type")
	}
}
