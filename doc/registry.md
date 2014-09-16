# Overview

The registry contains all the information needed to fetch and install packages. At the very
minimum it contains: the software version, installer URL and installer type.

Its canonical location is: <https://raw.github.com/lvillani/just-install/master/just-install.json>

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT",
"RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC
2119](http://tools.ietf.org/html/rfc2119).


## Structure

The registry is a [JSON](http://json.org/) file with comments starting with the hash (`#`) symbol
and extending until the end of the line. Since strict JSON doesn't support comments, these SHOULD
be stripped before handing off the registry to the JSON parser.

The minimal registry file is a JSON object containing the "version" key which refers to the schema
version information and the "package" key mapping to an empty JSON object.

```json
{
    "version": 2,
    "packages": {}
}
```

### Schema Version

The schema version is REQUIRED and MUST be represented with the "version" key mapping to a
non-negative integer value greater or equal than 2 in the root JSON object. For historical
reasons, version 1 refers to the old YAML-based format (called "The Catalog").

### Package Information

Package information is contained within a REQUIRED JSON object denoted by the key "packages".

Inside said JSON object there MAY be one or more packages whose keys map to JSON objects described
below. The keys are strings which denote the package's name as shown to users at the command line
interface. For example, a package with name "python27" can be installed by running
`just-install.exe python27`.

Package names SHALL be ordered alphabetically, to keep the file easier to read and modify.

The JSON object associated to a package name MUST contain the following mappings:

- __version__: REQUIRED - A string containing the program's version.
- __installer__: REQUIRED - A JSON object which contains the following mappings:
  + __kind__: REQUIRED - One of:
    - "advancedinstaller"
    - "as-is"
    - "conemu"
    - "custom"
    - "easy_install_26"
    - "easy_install_27"
    - "innosetup"
    - "msi"
    - "nsis"
    - "zip"
  + __x86__: REQUIRED - URL string to the installer for x86 architecture. The URL MAY contain the
    literal string `${version}` which will be expanded with the value of the `version` key in the
    parent JSON object precedently described.
  + __x86_64__: OPTIONAL - URL string to the installer for x86_64 architecture. Variable expansion
    is the same as above. In case this key is missing, `just-install` automatically falls back to
    the `x86` entry.
  + __options__: OPTIONAL - A JSON object which is passed to the handler of a specific package
    type and its interpretation is implementation-defined.

What follows is a minimal example of a valid registry file with two packages:

```json
{
    "version": 2,
    "packages": {
        "python2.7": {
            "version": "2.7.8",
            "installer": {
                "kind": "msi",
                "x86": "https://www.python.org/ftp/python/${version}/python-${version}.msi",
                "x86_64": "https://www.python.org/ftp/python/${version}/python-${version}.amd64.msi"
            }
        },
        "python2.7-pip": {
            "version": "latest",
            "installer": {
                "kind": "custom",
                "x86": "https://raw.github.com/pypa/pip/master/contrib/get-pip.py",
                "options": {
                    "arguments": ["\\Python27\\python.exe", "${installer}"]
                }
            },
        }
    }
}
```
