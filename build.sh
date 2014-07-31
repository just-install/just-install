#!/bin/sh

set -e

#
# Build executable
#

gox -osarch="windows/386" -output="just-install" .
upx -9 just-install.exe

#
# "Publish" just-install.exe to gh-pages repository
#

if [ "${1}" == "publish" ]; then
    mv just-install.exe ../just-install.exe
    git checkout gh-pages
    mv ../just-install.exe ./just-install.exe
    git commit -a --amend --no-edit
    git checkout master
fi
