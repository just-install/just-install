# Change Log

All notable changes to this project are documented in this file.

## 3.0.0 - Work in progress

### Breaking Changes

* Due to changes in the MSI authoring tool, uninstall the old version of just-install before
  upgrading.
* Changed the format of the registry file which is now documented and validated through the use of
  JSON schema. The way variable expansion is done was also changed with this release and we invite
  you to give a look at the new [just-install.json](just-install.json) file for an example of the
  new syntax.
* The `checklinks` and `self-update` actions have been removed.


### Added

* It is now possible to specify per-architecture install options. See the
  [perl](https://github.com/lvillani/just-install/blob/3ec45b3f03c01df68aa713269a3f0722019f81d5/just-install.json#L383-L388)
  entry as an example.


### Changed

* Switched from AdvancedInstaller back to WiX as the tool used to generate just-install's MSI
  package. This will probably cause just-install to appear twice in "Add/Remove programs" for some
  users.




## 2.3.1 - 2015-02-03

### Changed

* Architecture detection now uses the `%ProgramFiles(x86)%` environment variable to determine
  whether Windows is 32-bit or 64-bit capable. This should also fix a problem where just-install
  failed to start on 32-bit platforms.
* The `${version}` variable is now expanded also during shim creation.




## 2.3.0 - 2014-12-07

### Added

* A new testing infrastructure now ensures that all installers are still reachable after
  each commit.

### Changed

* Prompt users to upgrade in case the registry file format has changed in a non-backward-compatible
  way.
* Shim executables are now created using exeproxy, which replaces the old "mklink" way. You may want
  to refresh the shims by calling "just-install -s [pkg...]"




## 2.2.0 - 2014-09-21

### Added

* Added a new `extension` registry option to specify a custom extension to be appendend to
  downloaded files.
* The `%ProgramFiles%` and `%ProgramFiles(x86)%` environment variables get normalized at startup
  according to the scheme described in
  [bug #47](https://github.com/lvillani/just-install/issues/47)
* Some executables are symlinked to `%SystemDrive%\just-install` (only on Windows Vista and later).
* Added a new command-line `-s` switch to force regeneration of shim executables without having to
  re-install the program again. E.g.: `just-install -s mercurial`.


### Changed

* just-install now comes as an MSI package.
* `just-install self-update` is now an alias for `just-install -f just-install`.
* Installers and executables in general are now launched directly instead of going through the
  shell.
* just-install now honors registry entries not having the installer as first argument for entries
  of "custom" type.


### Removed

* just-install will no longer try to copy itself to `%WinDir%`.




## 2.1.0 - 2014-08-22

### Added

* Add support for wrapped installers (e.g. MSI file in a ZIP container). To see how to use this
  feature check out the "colemak" and "smartkey" entries from the registry.
* Add support for extracting ZIP files to an arbitrary location on disk. To see how this feature
  works, see the "depends" and "sysinternals" entries from the registry.

### Fixed

* Now honoring the "arguments" array for "custom" installers.
* Just-Install exits with an error if it fails to parse the registry.




## 2.0.1 - 2014-08-12

* Embedded manifest to require elevation.
* Embedded the icon in the executable again.




## 2.0.0 - 2014-08-5

* New command line interface, run `just-install --help` for help.
* Does not require the Visual C++ 2008 Runtime to be installed anymore.
* Antivirus program should not flag just-install as a virus anymore.
* More solid self-update functionality.
* New catalog file format.
* Rewritten in Go.
