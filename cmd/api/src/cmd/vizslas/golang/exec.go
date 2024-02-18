package golang

import (
	"os"
	"os/exec"
)

func PushDirectory(path string, delegate func() error) error {
	if previousDir, err := os.Getwd(); err != nil {
		return err
	} else if err := os.Chdir(path); err != nil {
		return err
	} else {
		defer os.Chdir(previousDir)

		return delegate()
	}

}

func Exec(cmd []string) ([]byte, error) {
	return exec.Command(cmd[0], cmd[1:]...).Output()
}

