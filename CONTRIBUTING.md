# Submitting New Catalog Entries

Fork this repository, make your changes to `just-install.json`, then submit a pull request.

Guidelines to follow:

* Read [this (rough) description of the registry file](doc/registry.md)
* Registry entries are listed in alphabetical order, make sure yours fit with this scheme.
* Prefer unversioned entries: if a software provides an URL to always get the latest installer
  try to use that.
* When an installer combines both 32-bit and 64-bit versions of an application, only add the
  required `x86` entry. That is: add both `x86` and `x86_64` URLs only when they actually differ.
* For development libraries: prefer ones compiled with Visual Studio since it's the native
  toolchain on Windows.

# How To Test Changes Locally

* To test the latest features, you can install the Beta version that gets automatically rebuilt after each commit on the master branch by running `msiexec /i http://beta.just-install.it`
* Run `just-install update` to fetch the latest version of the registry file.
* The registry is saved at `%TEMP%\just-install\just-install-v3.json`
* Make your changes to said file:
  - If you have to update an entry, just change it.
  - If you have to add a new entry, start with a similar one already in the file by copying and
    pasting it.
* Test your changes locally by running `just-install <entry name>`.
* Now copy the file over `just-install.json` in your Git repository and submit a PR.
