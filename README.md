# Just Install

<img src="misc/cube.svg" align="right" width="200" height="200"/>

_The simple package installer for Windows_

[![Build status](https://github.com/just-install/just-install/workflows/CI/badge.svg)](https://github.com/just-install/just-install/actions?query=workflow%3ACI)
[![License](https://img.shields.io/badge/license-GPL%203.0-blue.svg?style=flat)](https://choosealicense.com/licenses/gpl-3.0/)
[![Semver](https://img.shields.io/badge/version-v3.4.6-blue.svg?style=flat)](https://github.com/just-install/just-install/blob/master/CHANGELOG.md)

---

just-install is a simple program which automates software installation on Windows. It tries to
do one simple thing and do it well: download a `setup.exe` and install it, silently.


## Installation

For up-to-date install instructions, please visit <https://just-install.github.io>.


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
