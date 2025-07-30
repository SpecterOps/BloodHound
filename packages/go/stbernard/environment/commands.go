package environment

import (
	"fmt"
	"os"
	"os/exec"
)


// GoCommand will attempt to locate the preferred go binary
func GoCommand() string {
	// Allow users to override go root location
	goRoot := os.Getenv("GOROOT")
	if len(goRoot) != 0 {
		goCmd := fmt.Sprintf("%s/bin/go", goRoot)
		_, err := exec.LookPath(goCmd)
		if err == nil {
			return goCmd
		}
	}
	_, err := exec.LookPath("go")
	if err == nil {
		return "go"
	}
	panic("go does not seem to be installed!")
}


// YarnCommand checks which yarn executable is in the path
func YarnCommand() string {
	// Allow users to override YARN command location
	yarnCmd := os.Getenv("ST_YARNCMD")
	if len(yarnCmd) != 0 {
		_, err := exec.LookPath(yarnCmd)
		if err == nil {
			return yarnCmd
		}
	}
	// Arch, etc.
	_, err := exec.LookPath("yarn")
	if err == nil {
		return "yarn"
	}
	// Debian, etc. (yarn conflicts with other software)
	_, err = exec.LookPath("yarnpkg")
	if err == nil {
		return "yarnpkg"
	}
	panic("yarn does not seem to be installed!")
}
