The Catalog
===========

Overview
--------

The catalog contains all the information `just-install` needs to fetch and
install packages. At the very minimum it contains: the software version,
installer URL and installer type (so that `just-install` knows how to perform a
silent installation.)



File Format
-----------

The official catalog file is a YAML document residing at
<https://raw.github.com/lvillani/just-install/master/catalog/catalog.yml>.

Top level entries are package names which, in YAML parlance, are keys of an
associative array data type. The data they map to is another associative array
which contains all metadata needed to fetch and install the package, the
structure of which is described below.

## Metadata

### Required Entries

* `version`: Designates the version of the package, it can be a number, a
  string, anything.
* `url`: Direct URL to the installation program or MSI/ZIP archive.
* `type`: The installer type, which is one of the following:
  + `as-is`: Runs the installer as-is;
  + `conemu`: Specific to [ConEmu](http://code.google.com/p/conemu-maximus5/);
  + `innosetup`: An installer created with [InnoSetup](http://www.jrsoftware.org/isinfo.php);
  + `msi`: A [Windows Installer](http://msdn.microsoft.com/en-us/library/windows/desktop/cc185688%28v=vs.85%29.aspx) MSI package;
  + `nsis`: An installer created with [NSIS](http://nsis.sourceforge.net/Main_Page);
  + `zip`: A ZIP archive, which will be extracted to `%SystemDrive`;

### Optional Entries

* `path`: A list of entries to add to `%SystemDrive\rc.cmd` (a batch script to
  setup the `%PATH%` environment variable for console users), such entries are
  added if _and only if_ they're not there already.
