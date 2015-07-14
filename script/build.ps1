$BETA = $TRUE
$HERE = Split-Path -Parent $MyInvocation.MyCommand.Definition
$TOP_LEVEL = Split-Path -Parent $HERE

$env:JustInstallVersion = "3.0.0"

cd $TOP_LEVEL

#
# Clean
#

rm -Force *.exe
rm -Force *.msi
rm -Force *.wixobj
rm -Force *.wixpdb

#
# Build
#

godep go build -o just-install.exe bin\just-install.go

#
# Build MSI
#

candle just-install.wxs
light just-install.wixobj

if ($BETA) {
	mv just-install.msi just-install-beta.msi
}

#
# Upload MSI
#

if (-Not (Test-Path "..\just-install-web")) {
	pushd ..
		git clone -b gh-pages ssh://git@github.com/lvillani/just-install.git just-install-web
	popd
}

mv -Force *.msi ..\just-install-web

pushd ..\just-install-web
	git status
	git add *.msi
	git commit -a --amend --no-edit
	git push -f origin gh-pages
popd
