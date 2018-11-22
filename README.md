# Just Install

<img src="https://cdn.rawgit.com/just-install/just-install/4953fdccb9614bbdb2b77991610db6b99b1757d1/misc/cube.svg" align="right" width="200" height="200"/>

_The simple package installer for Windows_

[![Build status](https://ci.appveyor.com/api/projects/status/wpof4badsg7y0l3s/branch/master?svg=true)](https://ci.appveyor.com/project/lvillani/just-install/branch/master)
[![License](https://img.shields.io/badge/license-GPL%203.0-blue.svg?style=flat)](https://choosealicense.com/licenses/gpl-3.0/)
[![Semver](https://img.shields.io/badge/version-v3.4.2-blue.svg?style=flat)](https://github.com/just-install/just-install/blob/master/CHANGELOG.md)

---

:question: **Looking for help?** We recently decommissioned our Gitter rooms. If you want to ask a
question, please open an issue by clicking
[here](https://github.com/just-install/helpdesk/issues/new).

---

just-install is a simple program which automates software installation on Windows. [Unlike the
alternatives](https://lorenzo.villani.me/2013/04/08/just-install-my-stuff/), we strive to do one
simple thing and do it well: download a `setup.exe` and install it, without bothering the user.

To see the list of available packages head over to <https://just-install.it>.


## Installation

Run this command in a command prompt, as an Administrator:

```batch
msiexec.exe /i https://go.just-install.it
```

If you would like a more traditional means of installation then download
[just-install.msi](https://go.just-install.it) and double click the file to install it yourself.

If you would like to automatically install programs when `just-install.exe` is launched, use the
customizer [here](https://just-install.it/customizer.html).

If you want to try the next upcoming version of just-install, then run the following:

```batch
msiexec.exe /i https://unstable.just-install.it
```

## Quick start

To install a package, for example Firefox, run:

    just-install firefox

There are also other commands and flags that are described in the output of `just-install help`.


## Development

To contribute a new package, see
[here](https://github.com/just-install/registry/blob/master/README.md).

To work on just-install itself you will need Git, the Go compiler and the WiX Toolset. You can
install them with just-install itself:

    just-install git go wix


## Credits

The cube icon is derived from the one available from [Ionicons](https://ionicons.com/).
