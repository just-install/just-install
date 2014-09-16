A brief list of reasons for starting this project are available here:
<http://lorenzo.villani.me/2013/04/08/just-install-my-stuff/>

* Support for any x86 and x86_64 version of Windows still being supported by Microsoft,
  excluding "ExtendedSupport" releases.
  + Windows Vista
  + Windows 7
  + Windows 8
  + Windows 8.1
* Installing `just-install` must require copying its stand-alone executable somewhere
  on the target system. No installer or complex script needed.
* No other pre-requisites besides the operating system (no .NET runtime, or other runtime
  necessary)
* Just download and run the installer (silently if possible). Do not ovverride installation
  paths or handle upgrades. This is the responsibility of the installer itself. If it
  is bugged, file a complaint with the upstream developer.
* Do not alter the system in any way except for storing a local copy of the catalog and
  downloaded installers.
* Keep `just-install` as dumb as possible.