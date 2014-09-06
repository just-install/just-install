# Change Log

All notable changes to this project are documented in this file.


## 2.2.0 - Work in progress

### Changed

* just-install now comes as an MSI package.
* just-install will no longer try to copy itself to `%WinDir%`.
* `just-install self-update` is now an alias for `just-install -f just-install`.


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
