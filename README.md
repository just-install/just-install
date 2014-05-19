Just Install
============

[![Gittip](http://img.shields.io/gittip/lvillani.svg)](https://www.gittip.com/lvillani/)

![Box](https://raw.githubusercontent.com/lvillani/just-install/master/box.png)

*The stupid package installer for Windows*

Chocolatey, Ninite and Npackd do way too much. I 
[needed something simple](http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/) to
install stuff on Windows machines, here it is.




Installation
------------

Download [just-install.exe](http://lvillani.github.io/just-install/just-install.exe)
and double click it. Boom! `just-install` is now in your `%PATH%`. How easy is
that?

In a hurry? Here's a mnemonic link you can use to bootstrap a new machine: <http://go.just-install.it>

Feeling geeky? Copy and paste this line in a PowerShell console, then double click `just-install.exe` on
your Desktop.

    (New-Object System.Net.WebClient).DownloadFile("http://is.gd/justinstall", "${env:UserProfile}\Desktop\just-install.exe")




Usage
-----

Open an "Administrative Console Prompt" (that is: run "cmd.exe" as an Administrator) and type:

    just-install -h




Credits
-------

Box icon designed by Даниил Пронин from the Noun Project - Creative Commons – Attribution (CC BY 3.0)
