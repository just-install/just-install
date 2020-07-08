# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 3.4.8 - 2020-07-08

### Added

- Added support for `appx` and `jetbrains-nsis` installer types.

## 3.4.7 - 2019-12-21

### Changes

- The audit command now skips EitherMouse, since it does weird stuff from GitHub Actions
  infrastructure.

## 3.4.6 - 2019-11-30

### Changes

- Failure to install a package no longer prevents installing remaining packages. The application
  still exits with a failure status code.

## 3.4.5 - 2018-01-20

### Changes

- Push stable releases to GitHub Pages in addition to GitHub releases. Changed install command shown
  on registry version mismatch to reflect this change.

## 3.4.4 - 2018-12-23

### Changes

- Stable releases are now published on GitHub.

## 3.4.3 - 2018-12-22

### Fixed

- Downloads interrupted with no apparent errors after about 10 seconds.

## 3.4.2 - 2018-11-22

### Changed

- 64-bit only packages are now supported.
- Default registry URL has been changed to
  <https://just-install.github.io/registry/just-install-v4.json> to prepare for our migration
  back to a github.io domain.

### Fixed

- `just-install --registry foo.json list` now lists packages from custom registries specified on the
  command line.

## 3.4.0 - 2017-06-04

### Changed

- The most user-visible change should be that most stuff should be now served via TLS.

## 3.3.1 - 2017-02-09

Fixes to the CI system to produce 32-bit executables and installer packages.

## 3.3.0 - 2017-01-04

### Added

- It is now possible to append command line arguments to the executable to create a self-installing
  binary that automatically installs the given package list.

### Changed

- just-install won't automatically load `just-install.json` from the current directory anymore.
  Use the `-r, --registry` command line flag to specify an alternate path.

## 3.2.0 - 2016-10-28

### Changed

- Reliability improvements.

## 3.1.0 - 2015-09-06

### Addded

- Add a `--download-only` command line switch to download installers without installing them.

### Changed

- We now download the registry and all installers under `%TEMP%\just-install`.

## 3.0.0 - 2015-07-25

### Breaking Changes

- Due to changes in the MSI authoring tool, uninstall the old version of just-install before
  upgrading by running `msiexec /x {CEB24764-4726-4E3D-A66B-BE75ABAC4CD8} /qn`
- Changed the format of the registry file which is now documented and validated through the use of
  JSON schema. The way variable expansion is done was also changed with this release and we invite
  you to give a look at the new [just-install.json](just-install.json) file for an example of the
  new syntax.
- We now create shims under `%SystemDrive%\Shims`. The old path (`%SystemDrive%\just-install`) won't
  be migrated automatically so that your old shims will still work with your actual setup. We
  display a prominent warning until you remove the old `%SystemDrive%\just-install` directory. You
  can safely copy your old shims to their new place without having to regenerate them.
- The `checklinks` and `self-update` actions have been removed.

### Added

- It is now possible to specify per-architecture install options. See the
  [perl](https://github.com/just-install/just-install/blob/3ec45b3f03c01df68aa713269a3f0722019f81d5/just-install.json#L383-L388)
  entry as an example.

### Changed

- Switched from AdvancedInstaller back to WiX as the tool used to generate just-install's MSI
  package. This will probably cause just-install to appear twice in "Add/Remove programs" for some
  users.
- The registry is automatically updated if the locally cached copy is more than 24 hours old, which
  means that, in most cases, you don't have to manually run `just-install update` anymore.

### Fixed

- The devious owners of `amd.com`, `codeplex.com` and `java.oracle.com` put some lame safeguards
  against hot-linking, which broke downloads from said websites. We have implemented some
  workarounds to get downloads working again.

## 2.3.1 - 2015-02-03

### Changed

- Architecture detection now uses the `%ProgramFiles(x86)%` environment variable to determine
  whether Windows is 32-bit or 64-bit capable. This should also fix a problem where just-install
  failed to start on 32-bit platforms.
- The `${version}` variable is now expanded also during shim creation.

## 2.3.0 - 2014-12-07

### Added

- A new testing infrastructure now ensures that all installers are still reachable after
  each commit.

### Changed

- Prompt users to upgrade in case the registry file format has changed in a non-backward-compatible
  way.
- Shim executables are now created using exeproxy, which replaces the old "mklink" way. You may want
  to refresh the shims by calling "just-install -s [pkg...]"

## 2.2.0 - 2014-09-21

### Added

- Added a new `extension` registry option to specify a custom extension to be appended to
  downloaded files.
- The `%ProgramFiles%` and `%ProgramFiles(x86)%` environment variables get normalized at startup
  according to the scheme described in
  [bug #47](https://github.com/just-install/just-install/issues/47)
- Some executables are symlinked to `%SystemDrive%\just-install` (only on Windows Vista and later).
- Added a new command-line `-s` switch to force regeneration of shim executables without having to
  re-install the program again. E.g.: `just-install -s mercurial`.

### Changed

- just-install now comes as an MSI package.
- `just-install self-update` is now an alias for `just-install -f just-install`.
- Installers and executables in general are now launched directly instead of going through the
  shell.
- just-install now honors registry entries not having the installer as first argument for entries
  of "custom" type.

### Removed

- just-install will no longer try to copy itself to `%WinDir%`.

## 2.1.0 - 2014-08-22

### Added

- Add support for wrapped installers (e.g. MSI file in a ZIP container). To see how to use this
  feature check out the "colemak" and "smartkey" entries from the registry.
- Add support for extracting ZIP files to an arbitrary location on disk. To see how this feature
  works, see the "depends" and "sysinternals" entries from the registry.

### Fixed

- Now honoring the "arguments" array for "custom" installers.
- Just-Install exits with an error if it fails to parse the registry.

## 2.0.1 - 2014-08-12

- Embedded manifest to require elevation.
- Embedded the icon in the executable again.

## 2.0.0 - 2014-08-5

- New command line interface, run `just-install --help` for help.
- Does not require the Visual C++ 2008 Runtime to be installed anymore.
- Antivirus program should not flag just-install as a virus anymore.
- More solid self-update functionality.
- New catalog file format.
- Rewritten in Go.
