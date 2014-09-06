# Just Install

<img src="https://cdn.rawgit.com/lvillani/just-install/4953fdccb9614bbdb2b77991610db6b99b1757d1/misc/cube.svg" align="right" width="200" height="200"/>

_The stupid package installer for Windows_

[![Semver](http://img.shields.io/badge/semver-v2.0.1-blue.svg)](https://github.com/lvillani/just-install/releases)
[![Gittip](http://img.shields.io/gittip/lvillani.svg)](https://www.gittip.com/lvillani/)
[![License](http://img.shields.io/badge/license-GPL%203.0-blue.svg)](http://choosealicense.com/licenses/gpl-3.0/)
[![Stories in Ready](https://badge.waffle.io/lvillani/just-install.png?label=ready&title=Ready)](https://waffle.io/lvillani/just-install)

--------------------------------------------------------------------------------

Chocolatey, Ninite and Npackd are way too complicated. I
[needed something simple](http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/) to install
stuff on Windows machines, here it is.

And when I say stupid, I really mean it. It is so dumb it cannot even handle errors! If one
occurs, just-install will happily spit an undecipherable stack trace on the console. The only
thing it is capable of is downloading a setup program and silently execute it. This simplicity
means that it's trivial to add support for new software, seriously,
[check out the registry](https://github.com/lvillani/just-install/blob/master/just-install.json)!


## Installation

Run this command in a command prompt:

```batch
msiexec.exe /i http://go.just-install.it
```

If you like more traditionals means of installation then download
[just-install.msi](http://go.just-install.it), then double click it to install it yourself.


## Removal

You can uninstall `just-install` from Windows' Control Panel.


## Usage

    NAME:
       just-install - The stupid package installer for Windows

    USAGE:
       just-install [global options] command [command options] [arguments...]

    VERSION:
       2.1.0

    COMMANDS:
       list     List all known packages
       self-update  Update just-install itself
       update   Update the registry
       help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --arch, -a       Force installation for a specific architecture (if supported by the host).
       --force, -f      Force package re-download
       --help, -h       show help
       --version, -v    print the version


## Credits

Cube icon derived from the one available from [Ionicons](http://ionicons.com/).


## Donating

Support this project and [others by Lorenzo Villani](https://github.com/lvillani/) via
[gittip](https://www.gittip.com/lvillani/).

[![Support via Gittip](https://cdn.rawgit.com/lvillani/gittip-badge/v1.0.0/dist/gittip.svg)](https://www.gittip.com/lvillani/)
