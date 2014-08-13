Just Install
============

<img align="right" src="misc/cube.png" width="200" height="200"/>

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




Installation
------------

Download [just-install.exe](http://lvillani.github.io/just-install/just-install.exe) and double
click it. Boom! `just-install` is now in your `%PATH%`. How easy is that?

In a hurry? Here's a mnemonic link you can use to bootstrap a new machine:
<http://go.just-install.it>

Feeling geeky? Copy and paste this line in a PowerShell console, then double click
`just-install.exe` on your Desktop.

```posh
(New-Object System.Net.WebClient).DownloadFile("http://go.just-install.it", "${env:UserProfile}\\Desktop\\just-install.exe")
```



Removal
-------

Remember when you double clicked on `just-install.exe` and magically found it in your `%PATH%`?
That's because it copied itself to `%WINDIR%`.

So, to completely remove `just-install` from your system, simply delete
`%WINDIR%\just-install.exe` and `%TEMP%\just-install.json`. You might also have
`just-install.old.exe` lying around (if you used the self-update function) so better delete it
too. Run these commands from within `cmd.exe`:

```bat
del /Q %WINDIR%\just-install.exe
del /Q %TEMP%\just-install.json
```




Usage
-----

    NAME:
       just-install - The stupid package installer for Windows

    USAGE:
       just-install [global options] command [command options] [arguments...]

    VERSION:
       2.0.0

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




Credits
-------

Cube icon derived from the one available from [Ionicons](http://ionicons.com/).
