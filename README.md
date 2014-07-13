Just Install
============

<img align="right" src="misc/cube.png" width="200" height="200"/>

*The stupid package installer for Windows*

Chocolatey, Ninite and Npackd are way too complicated. I [needed something
simple](http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/) to install stuff on Windows
machines, here it is.

And when I say stupid, I really mean it. It is so dumb it cannot even handle errors! If one occurs,
just-install will happily spit an undecipherable stack trace on the console. The only thing it is
capable of is downloading a setup program and silently execute it. This simplicity means that it's
trivial to add support for new software, seriously, [check out the
catalog](https://github.com/lvillani/just-install/blob/master/catalog/catalog.yml)!




Installation
------------

Download [just-install.exe](http://lvillani.github.io/just-install/just-install.exe)
and double click it. Boom! `just-install` is now in your `%PATH%`. How easy is
that?

In a hurry? Here's a mnemonic link you can use to bootstrap a new machine: <http://go.just-install.it>

Feeling geeky? Copy and paste this line in a PowerShell console, then double click `just-install.exe` on
your Desktop.

```posh
(New-Object System.Net.WebClient).DownloadFile("http://go.just-install.it", "${env:UserProfile}\Desktop\just-install.exe")
```



Removal
-------

Remember when you double clicked on `just-install.exe` and magically found it in your `%PATH%`?
That's because it copied itself to `%WINDIR%`.

So, to completely remove `just-install` from your system, simply delete `%WINDIR%\just-install.exe`
and `%TEMP%\catalog.json`. You might also have `just-install.old.exe` lying around (if you used the
self-update function) so better delete it too. Run these commands from within `cmd.exe`:

```bat
del /Q %WINDIR%\just-install.exe
del /Q %WINDIR%\just-install.old.exe
del /Q %TEMP%\catalog.json
```



Usage
-----

Open an "Administrative Console Prompt" (that is: run "cmd.exe" as an Administrator) and type:

    just-install -h




Credits
-------

Cube icon derived from the one available from [Ionicons](http://ionicons.com/).
