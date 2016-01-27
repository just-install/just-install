# Submitting New Catalog Entries

Fork this repository, make your edit to `just-install.json` then submit a pull request.

Guidelines to follow:

- Catalog entries are listed in alphabetical order, make sure yours fit with this scheme.
- Prefer unversioned entries: if a software provides a way to always get the latest installer
  try to use it so that users always get the latest version.
- When an installer combines both 32-bit and 64-bit versions of an application, only add the
  required `x86` entry. That is: add both `x86` and `x86_64` URLs when they actually differ.
- Prefer the x86_64 edition of the software:
  + Unless there are several compatibility or support issues (e.g.: Python)
- For development libraries: prefer ones compiled with Visual Studio since it's the native
  toolchain on Windows.
