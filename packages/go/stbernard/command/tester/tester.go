package tester

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "test"
	Usage = "Run tests for entire workspace"
)

type Config struct {
	Environment environment.Environment
}

type command struct {
	config Config
}

func (s command) Usage() string {
	return Usage
}

func (s command) Name() string {
	return Name
}

func (s command) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing module absolute paths: %w", err)
	} else if jsPaths, err := workspace.ParseJSAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing JS absolute paths: %w", err)
	} else {
		fmt.Println(modPaths)
		fmt.Println(jsPaths)
		return nil
	}
}

func Create(config Config) (command, error) {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[2:]); err != nil {
		cmd.Usage()
		return command{}, fmt.Errorf("parsing %s command: %w", Name, err)
	} else {
		return command{config: config}, nil
	}
}
