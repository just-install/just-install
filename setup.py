from distutils.core import Command, setup
from subprocess import check_call


#
# Helper commands
#

class FreezeCommand(Command):
    description = "Creates a stand-alone executable"
    user_options = []

    def initialize_options(self):
        pass

    def finalize_options(self):
        pass

    def run(self):
        check_call(["pyinstaller.exe", "-c", "-F", "just-install.py"])


class InstallerCommand(Command):
    description = "Creates a distributable package"
    user_options = []

    def initialize_options(self):
        pass

    def finalize_options(self):
        pass

    def run(self):
        check_call(["iscc.exe", "just-install.iss"])


#
# Setup
#

setup(
    name="JustInstall",
    version="1.0.0",
    license="GNU General Public License version 3",
    scripts=["just-install.py"],
    cmdclass={
        "freeze": FreezeCommand,
        "installer": InstallerCommand,
    }
)
