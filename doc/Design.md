Design
======

A brief list of reasons for starting this project are available here:
<http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/>

In general we:

* Support any **x86_64** version of Windows which is still being supported by
  Microsoft itself (excluding releases in "Extended Support").
* Make `just-install` installable without any kind of pre-requisite (i.e.: no
  separate .NET runtime requirement, or anything similar) besides Windows
  itself.
* Never, ever try to outsmart the installer package by overriding installation
  path and options.
* Never ever try to fiddle with system-wide settings such as `PATH` variable
  manipulation. While we agree that making software available to system path is
  useful, the preferred way is to provide an opt-in mechanism to achieve that
  goal (in the form of a `cmd.exe` startup script, at the moment).
* Try to keep `just-install` as simple and as dumb as possible.
