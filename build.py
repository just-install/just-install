#!/usr/bin/env python2.7
#
# just-install - The stupid package installer
#
# Copyright (C) 2013-2016  Lorenzo Villani
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 3 of the License.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

from __future__ import print_function

import glob
import os
import shutil
import sys
from subprocess import check_output as get_output
from subprocess import check_call

HERE = os.path.dirname(__file__)

if sys.platform == "win32":
    EXE = "just-install.exe"
else:
    EXE = "just-install"


def main():
    setup()
    clean()
    build()

    if sys.platform == "win32":
        build_msi()

    if "CI" in os.environ:
        appveyor()


def setup():
    call("go", "get", "github.com/kardianos/govendor")
    call("govendor", "sync")


def clean():
    def remove(*args):
        for f in args:
            try:
                os.remove(f)
            except:
                pass

    remove("just-install")
    remove(*glob.glob("*.exe"))
    remove(*glob.glob("*.msi"))
    remove(*glob.glob("*.wixobj"))
    remove(*glob.glob("*.wixpdb"))


def build():
    version = get_output(["git", "describe", "--tags"])

    os.environ["JustInstallVersion"] = version[1:6]

    if "--skip-tests" not in sys.argv:
        call("go", "test", "-v")

    call("go", "build", "-o", EXE, "-ldflags", "-X main.version=" + version, "./bin")


def build_msi():
    call("candle", "just-install.wxs")
    call("light", "just-install.wixobj")


def appveyor():
    if "APPVEYOR_REPO_TAG_NAME" not in os.environ:
        return

    tag = os.environ["APPVEYOR_REPO_TAG_NAME"]

    shutil.move("just-install.exe", "just-install-{}.exe".format(tag))
    shutil.move("just-install.msi", "just-install-{}.msi".format(tag))


def call(*args):
    print("+", " ".join(args))
    check_call(args)


if __name__ == "__main__":
    main()
