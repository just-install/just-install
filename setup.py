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
        check_call(["pyinstaller.exe", "--clean", "--distpath=.", "-i", "box.ico", "-c", "-F", "-y", "just-install.py"])


#
# Setup
#

just_install = __import__("just-install")

setup(
    name="JustInstall",
    version=just_install.__version__,
    license="GNU General Public License version 3",
    scripts=["just-install.py"],
    cmdclass={
        "freeze": FreezeCommand,
    }
)
