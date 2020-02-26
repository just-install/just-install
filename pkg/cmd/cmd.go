// just-install - The simple package installer for Windows
// Copyright (C) 2020 just-install authors.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

// Run runs a command, printing the command line to standard output. Additional output is printed in
// case we run msiexec and it returns with code 3010 (short for "reboot needed").
func Run(args ...string) error {
	if len(args) < 1 {
		return errors.New("empty command line")
	}

	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	log.Println("Running", strings.Join(args, " "))

	err := cmd.Start()
	if err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if !ok {
			return err
		}

		status, ok := exiterr.Sys().(syscall.WaitStatus)
		if !ok {
			return err
		}

		// msiexec returns 3010 if install needs reboot later
		if strings.Contains(args[0], "msiexec") && status.ExitStatus() == 3010 {
			log.Printf("msiexec exited with code 3010, a reboot is required to complete installation")
			return nil
		}

		return err
	}

	return nil
}
