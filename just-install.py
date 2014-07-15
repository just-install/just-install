#!/usr/bin/env python
# -*- coding: utf-8 -*-
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


__version__ = "1.2.0"


import argparse
import os
import os.path
import platform
import shutil
import string
import subprocess
import sys
import tempfile
import time
import urllib
import urlparse
import win32process
import yaml
import zipfile


TEMP_DIR = tempfile.gettempdir()

CATALOG_URL = "http://raw.github.com/lvillani/just-install/master/catalog/catalog.yml"
CATALOG_FILE = os.path.join(TEMP_DIR, os.path.basename(CATALOG_URL))
CATALOG_LOCAL = os.path.join(os.path.dirname(__file__), "catalog", "catalog.yml")
CATALOG_SUPPORTED_VERSION = 1

DEFAULT_ARCH = 'x86_64' if platform.machine() == 'AMD64' else platform.machine()

SELF_INSTALL_PATH = os.path.join(os.environ['SystemRoot'], 'just-install.exe')
SELF_UPDATER_PATH = os.path.join(os.environ['SystemRoot'], 'just-install.old.exe')
SELF_UPDATER_URL = "http://github.com/lvillani/just-install/releases/download/latest/just-install.exe"


def main():
    args = parse_command_line_arguments()
    arch = args.arch

    if args.version:
        print("just-install v" + __version__)

    if args.update:
        update(args.updated_exe)
    else:
        maybe_auto_install()

    catalog = load_catalog(args.force)

    check_catalog_version(catalog)

    if args.list:
        for package in sorted(catalog.keys()):
            print "%s - %s" % (package.rjust(30), catalog[package]["version"])

    for package in args.packages:
        installer_type = catalog[package]["type"]
        installer_version = catalog[package]["version"]

        if arch in catalog[package]["installer"]:
            installer_url_template = string.Template(catalog[package]["installer"][arch])
        elif isinstance(catalog[package]["installer"], basestring):
            arch = ""
            installer_url_template = string.Template(catalog[package]["installer"])
        else:
            raise ValueError("%s: architecture not supported." % arch)

        installer_url = installer_url_template.substitute(version=installer_version)

        print "%s-%s %s" % (package, installer_version, arch)
        print "    Downloading ...  ",
        installer_path = download_file(installer_url, overwrite=args.force)

        print ""
        print "    Installing ..."
        install(installer_path, installer_type, catalog[package])


def maybe_auto_install():
    if not hasattr(sys, "frozen") or sys.argv[0] == SELF_UPDATER_PATH:
        return

    if sys.argv[0] != SELF_INSTALL_PATH:
        print "Self-installing ...  "
        shutil.copyfile(sys.argv[0], SELF_INSTALL_PATH)


def parse_command_line_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument("--updated-exe", help=argparse.SUPPRESS, nargs='?')  # Internal
    parser.add_argument("-a", "--arch", action="store", help="Enorce a specific architecture", default=DEFAULT_ARCH, type=str)
    parser.add_argument("-f", "--force", action="store_true", help="Always re-download files, including the catalog")
    parser.add_argument("-l", "--list", action="store_true", help="List packages available for installation")
    parser.add_argument("-u", "--update", action="store_true", help="Update just-install itself")
    parser.add_argument("-v", "--version", action="store_true", help="Show version")
    parser.add_argument("packages", help="Packages to install", type=str, nargs="*")

    return parser.parse_args()


def update(updated_exe):
    # We copy ourselves (better safe than sorry), download the new executable and re-launch
    # ourselves with the hidden --updated-exe flag. We wait a second to let Windows release file
    # locks and copy the updated exe in place. Users should se a console window flashing for a short
    # time. We have to do this since DETACHED_PROCESS doesn't give us stdout and we want to minimize
    # the amount of time we appear silent.
    if updated_exe:
        time.sleep(1)
        shutil.copyfile(updated_exe, SELF_INSTALL_PATH)
    else:
        print ""
        print "WARNING: You might see a console window flashing for a short time."
        print "         This is expected. Don't panic!"
        print ""
        print "Updating ...  ",
        downloaded = download_file(SELF_UPDATER_URL, overwrite=True)

        shutil.copyfile(SELF_INSTALL_PATH, SELF_UPDATER_PATH)
        subprocess.Popen([SELF_UPDATER_PATH, '-u', '--updated-exe', downloaded], creationflags=win32process.DETACHED_PROCESS)
        sys.exit(0)


def load_catalog(force_update):
    catalog_path = CATALOG_FILE

    if os.path.exists(CATALOG_LOCAL):
        catalog_path = CATALOG_LOCAL
    else:
        catalog_path = CATALOG_FILE

        if not os.path.exists(CATALOG_FILE) or force_update:
            print "Updating catalog ...  ",
            download_file(CATALOG_URL, overwrite=True)

    print "Loading catalog from " + catalog_path

    with open(catalog_path) as catalog:
        return yaml.load(catalog)


def check_catalog_version(catalog):
    v = catalog["catalog_version"]

    if not v == CATALOG_SUPPORTED_VERSION:
        print "This version of just-install does not support catalog version " + str(v)
        print "Please update just-install (run 'just-install -u') and try again."

        sys.exit(1)


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
        call(*[expand_installer_path(a, path) for a in env["custom_arguments"].split(" ")])
    elif kind.startswith("easy_install"):
        call(r'\Python' + kind[-2:] + '\Scripts\easy_install.exe', path)
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


def expand_installer_path(s, installer_path):
    return string.Template(s).substitute(installer=installer_path)


def call(*args):
    print('    (' + ' '.join(args) + ')')
    subprocess.check_call(args, shell=True)


def zip_extract(path, destination):
    zip_file = zipfile.ZipFile(path, "r")
    zip_file.extractall(destination)


if __name__ == '__main__':
    main()
