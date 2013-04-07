#!/bin/bash

set -eux

export PATH="${SYSTEMDRIVE}\Python27:${PATH}"
export PATH="${SYSTEMDRIVE}\pyinstaller-2.0:${PATH}"
export PATH="${WIX}\bin:${PATH}"

#
# Build standalone executable
#

python "${SYSTEMDRIVE}\pyinstaller-2.0\pyinstaller.py" -F just-install.py

#
# MSI
#

candle.exe just-install.wxi
light.exe  just-install.wixobj

#
# Upload MSI
#

git checkout gh-pages
git add just-install.msi
git commit -m 'Updated MSI'
git push
git checkout master
