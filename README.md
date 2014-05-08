Just Install
============

![Box](https://raw.githubusercontent.com/lvillani/just-install/master/box.png)

*The stupid package installer for Windows*

Chocolatey, Ninite, Npackd are way too slow, bloated and difficult to extend. I
[needed](http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/) a no-frills solution to
install stuff on Windows machines, here it is.




Installation
------------

Download [just-install.exe](http://lvillani.github.io/just-install/just-install.exe)
and double click it. Boom! `just-install` is now in your `%PATH%`. How easy is
that?

In a hurry? Here's a mnemonic link you can use to bootstrap a new machine, just type this in a
browser [is.gd/justinstall](http://is.gd/justinstall) then double-click the downloaded file. *"Is
(it) Good? Just Install!"*.

Are you on a fresh Windows Server install and only have the annoying Internet Explorer, with a bad
habit of blocking all downloads to give you a false sense of security? Don't want to fiddle with its
settings? Copy and paste this line in a PowerShell console, then double click `just-install.exe` on
your Desktop.

    (New-Object System.Net.WebClient).DownloadFile("http://is.gd/justinstall", "${env:UserProfile}\Desktop\just-install.exe")




Credits
-------

Box icon designed by Даниил Пронин from the Noun Project - Creative Commons – Attribution (CC BY 3.0)
