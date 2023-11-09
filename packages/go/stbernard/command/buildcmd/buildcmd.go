package buildcmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "build"
	Usage = "Build a list of targets"
)

type cmd struct{}

func (s cmd) Name() string {
	return Name
}

func (s cmd) Usage() string {
	return Usage
}

func (s cmd) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("could not find workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse module absolute paths: %w", err)
	} else if err := workspace.BuildMainPackages(cwd, modPaths); err != nil {
		return fmt.Errorf("could not build main packages: %w", err)
	} else {
		return nil
	}
}

func Create() (cmd, error) {
	var (
		command  = cmd{}
		buildCmd = flag.NewFlagSet(Name, flag.ExitOnError)
	)

	buildCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n", Usage, filepath.Base(os.Args[0]), Name)
	}

	if err := buildCmd.Parse(os.Args[2:]); err != nil {
		buildCmd.Usage()
		return command, fmt.Errorf("failed to parse %s command: %w", Name, err)
	} else {
		return command, nil
	}
}
