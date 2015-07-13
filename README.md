# Just Install

<img src="https://cdn.rawgit.com/lvillani/just-install/4953fdccb9614bbdb2b77991610db6b99b1757d1/misc/cube.svg" align="right" width="200" height="200"/>

_The stupid package installer for Windows_

[![Build Status](http://img.shields.io/travis/lvillani/just-install.svg?style=flat)](https://travis-ci.org/lvillani/just-install/)
[![License](http://img.shields.io/badge/license-GPL%203.0-blue.svg?style=flat)](http://choosealicense.com/licenses/gpl-3.0/)
[![Semver](http://img.shields.io/badge/version-v2.3.1-blue.svg?style=flat)](https://github.com/lvillani/just-install/blob/master/CHANGELOG.md)
[![ML](https://img.shields.io/badge/ML-justinstall@librelist.com-orange.svg)](http://librelist.com/browser/justinstall)

--------------------------------------------------------------------------------

**NEW**: Subscribe to the mailing list by sending an email to <justinstall@librelist.com>

Chocolatey, Ninite and Npackd are way too complicated. I
[needed something simple](http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/) to install
stuff on Windows machines, here it is.

And when I say stupid, I really mean it. It is so dumb it cannot even handle errors! If one
occurs, just-install will happily spit an undecipherable stack trace on the console. The only
thing it is capable of is downloading a setup program and silently execute it. This simplicity
means that it's trivial to add support for new software, seriously,
[check out the registry](https://github.com/lvillani/just-install/blob/master/just-install.json)!


## Installation

Run this command in a command prompt, as Administrator:

```batch
msiexec.exe /i http://go.just-install.it
```

If you like more traditional means of installation then download
[just-install.msi](http://go.just-install.it), then double click the file to install it yourself.


## Usage

To install a package:

    just-install firefox

To view a list of available packages:

    just-install list

To update the list of available packages:

    just-install update

To forcibly re-download an installer and re-run it:

    just-install -f firefox

To force installation of a package for a specific architecture (use "x86" or "x86_64"):

    just-install -a x86 go

In case you are lost, help is always few keystrokes away:

    just-install --help


## Contributing Packages

See [CONTRIBUTING.md](CONTRIBUTING.md)


## Windows 10 and Beyond

It looks like Microsoft is adding a package manager called
[OneGet](https://github.com/OneGet/oneget) to Windows 10. Since the ultimate purpose of just-install
is to be replaced by an official solution from Microsoft, we __won't__ officially support Windows
10, and we invite you to use and contribute to the development of OneGet so that the next release of
Windows gets a decent package manager.

We will continue to support just-install on Windows XP, 7, 8 and 8.1.


## Credits

Cube icon derived from the one available from [Ionicons](http://ionicons.com/).
