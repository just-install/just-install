#!/bin/sh

set -e

mv just-install.exe ../just-install.exe
git checkout gh-pages
mv ../just-install.exe ./just-install.exe
git commit -a --amend --no-edit
git push origin gh-pages -f
git checkout master
