#!/usr/bin/env python
#
# just-install - The stupid package installer
#
# Copyright (C) 2013, 2014  Lorenzo Villani
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#


__version__ = "1.1.0"


import argparse
import os
import os.path
import platform
import shutil
import string
import subprocess
import sys
import tempfile
import urllib
import urlparse
import yaml
import zipfile


TEMP_DIR = tempfile.gettempdir()
CATALOG_URL = "http://raw.github.com/lvillani/just-install/master/catalog/catalog.yml"
CATALOG_FILE = os.path.join(TEMP_DIR, os.path.basename(CATALOG_URL))
CATALOG_LOCAL = os.path.join(os.path.dirname(__file__), "catalog", "catalog.yml")
DEFAULT_ARCH = platform.machine()
SELF_INSTALL_PATH = os.path.join(os.environ['SystemRoot'], 'just-install.exe')
SELF_UPDATE_URL = "http://github.com/lvillani/just-install/releases/download/latest/just-install.exe"


def main():
    args = parse_command_line_arguments()
    arch = args.arch

    if args.version:
        print("just-install v" + __version__)

    if args.update:
        update()
    else:
        maybe_auto_install()

    fetch_catalog(args.force)

    catalog = load_catalog(CATALOG_FILE)

    if args.list:
        for package in sorted(catalog.keys()):
            print "%s - %s" % (package.rjust(30), catalog[package]["version"])

    for package in args.packages:
        installer_type = catalog[package]["type"]
        installer_version = catalog[package]["version"]

        if arch in catalog[package]["installer"]:
            installer_url_template = string.Template(catalog[package]["installer"][arch])
        elif isinstance(catalog[package]["installer"], basestring):
            installer_url_template = string.Template(catalog[package]["installer"])
        else:
            raise ValueError("%s: architecture not supported." % arch)

        installer_url = installer_url_template.substitute(version=installer_version)

        print "%s (%s)" % (package, installer_version)
        print "    Downloading ...  ",
        installer_path = download_file(installer_url, overwrite=args.force)

        print ""
        print "    Installing ..."
        install(installer_path, installer_type, catalog[package])


def maybe_auto_install():
    if not hasattr(sys, "frozen"):
        return

    sysroot = os.environ["SystemRoot"]

    if not sys.argv[0].startswith(sysroot):
        print "Self-installing ...  "
        shutil.copyfile(sys.argv[0], SELF_INSTALL_PATH)


def parse_command_line_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument("-a", "--arch", action="store", default=DEFAULT_ARCH, type=str)
    parser.add_argument("-f", "--force", action="store_true")
    parser.add_argument("-l", "--list", action="store_true")
    parser.add_argument("-u", "--update", action="store_true")
    parser.add_argument("-v", "--version", action="store_true")
    parser.add_argument("packages", type=str, nargs="*")

    return parser.parse_args()


def update():
    print "Self-updating ...  ",
    shutil.copyfile(download_file(SELF_UPDATE_URL, overwrite=True), SELF_INSTALL_PATH)


def fetch_catalog(force_update):
    if not hasattr(sys, "frozen") and os.path.exists(CATALOG_LOCAL):
        shutil.copyfile(CATALOG_LOCAL, CATALOG_FILE)
    elif not os.path.exists(CATALOG_FILE) or force_update:
        print "Updating catalog ...  ",
        download_file(CATALOG_URL, overwrite=True)
        print ""


def load_catalog(path):
    with open(path) as catalog:
        return yaml.load(catalog)


def download_file(url, overwrite=False):
    def progress_report(count, block_size, total_size):
        percent = int(count * block_size * 100 / total_size)

        # Sometimes, it goes over the top
        if percent > 100:
            percent = 100

        sys.stdout.write("%2d%%" % percent)
        sys.stdout.write("\b\b\b")
        sys.stdout.flush()

    basename = os.path.basename(urllib.unquote(urlparse.urlparse(url).path))
    destination = os.path.join(TEMP_DIR, basename)

    if overwrite or not os.path.exists(destination):
        urllib.urlretrieve(url, destination, reporthook=progress_report)

    return destination


def install(path, kind, env={}):
    if kind == "advancedinstaller":
        call(path, "/q", "/i")
    elif kind == "as-is":
        call(path)
    elif kind == "custom" and "custom_arguments" in env:
        call(path, *env["custom_arguments"].split(" "))
    elif kind == "innosetup":
        call(path, "/sp-", "/verysilent", "/norestart")
    elif kind == "microsoft":
        call(path, "/quiet", "/passive", "/norestart")
    elif kind == "msi":
        call("msiexec.exe", "/q", "/i", path, "REBOOT=ReallySuppress")
    elif kind == "nsis":
        call(path, "/S", "/NCRC")
    elif kind == "zip":
        zip_extract(path, os.environ['SystemDrive'] + '\\')
    else:
        raise TypeError("Unknown installer type: %s" % kind)


def call(*args):
    subprocess.check_call(args, shell=True)


def zip_extract(path, destination):
    zip_file = zipfile.ZipFile(path, "r")
    zip_file.extractall(destination)


if __name__ == '__main__':
    main()
