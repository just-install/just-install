Registry
========

The registry file is a JSON document with a single top-level JSON object that follows the schema
described in [just-install-schema.json](just-install-schema.json). This document roughly describes
the format of the registry file for humans :smile:

There are no examples in this document, the registry file itself is a living example of what you can
do.


## Top Level

The top-level JSON object must contain two keys:

* `version`: Contains the version of the registry file and is used to prompt upgrades from older
  versions of just-install. The version is bumped each time we make a backward-incompatible change
  to the file format.
* `packages`: This is a JSON object. Each key represents a package name and the value is itself a
  JSON object that contains the software version and instructions to get the installer. See "Package
  Entry" below for a description.


## Package Entry

Each entry is a JSON object that must contain at least the following two keys:

* `installer`: A JSON object that describes where the installer is and how to run it once
  downloaded. See "Installer Options" below for a description.
* `version`: The software's version. If you are adding an unversioned link that always points to the
  latest stable version use `latest` here.


## Installer

This JSON object must contain at least the following two keys:

* `x86`: The value is a string with the URL that must be used to download the installer. You can use
  `{{.version}}` as a placeholder for the package's version.
* `kind`: It can be one of the following:
  - `advancedinstaller`: Silently installs Advanced Installer packages;
  - `as-is`: Will just run the executable, as-is;
  - `copy`: Copy the file according to the `destination` parameter;
  - `custom`: Allows you to specify how to call the installer ([example](https://github.com/lvillani
    /just-install/blob/18876192c5ed7f24a3acaa34524d3680ec17da3e/just-install.json#L79-L101));
  - `easy_install26` and `easy_install_27`: used to install Python packages (the user must have
    installed Python 2.6 or Python 2.7 first);
  - `innosetup`: Silently installs InnoSetup packages;
  - `msi`: Silently installs Windows Installer packages;
  - `nsis`: Silently installs NSIS packages;
  - `zip`: [Runs](https://github.com/lvillani/just-
    install/blob/18876192c5ed7f24a3acaa34524d3680ec17da3e/just-install.json#L66-L78) an installer
    within a .zip file or [extracts](https://github.com/lvillani/just-
    install/blob/18876192c5ed7f24a3acaa34524d3680ec17da3e/just-install.json#L216-L231) it to a
    destination directory.


## Shims

Shims are a way to easily add executables to the `%PATH%`. They are created only if the user has
installed `exeproxy` (either through `just-install` itself or manually) and only if the package
entry specifies some shims to create.

Take, for example, the [Go entry](https://github.com/lvillani/just-
install/blob/18876192c5ed7f24a3acaa34524d3680ec17da3e/just-install.json#L336-L350): this will create
three executables called `go.exe`, `godoc.exe` and `gofmt.exe` under `%SystemDrive%\Shims` that will
forward any argument to the original file.

This way users don't have to add a directory for each installed software to their `%PATH%` since
they can just add `%SystemDrive%\Shims`.


## Placeholders

In some places you can use the following placeholders:

* `{{.version}}`: This placeholder gets expanded with the package's version.
* `{{.installer}}`: This placeholder gets replaced with the absolute path to the downloaded
  installer executable.
* `{{.ENV_VAR}}`: Where `ENV_VAR` is any environment variable found on the system. All environment
  variables are normalized to upper case so, for example, `%SystemDrive%` becomes available as
  `{{.SYSTEMDRIVE}}`. One exception is `%ProgramFiles(x86)%` that gets normalized as
  `{{.PROGRAMFILES_X86}}` (notice the lack of parentheses).
