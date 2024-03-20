package tester

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/yarn"
)

const (
	Name  = "test"
	Usage = "Run tests for entire workspace"
)

type command struct {
	env      environment.Environment
	yarnOnly bool
	goOnly   bool
}

func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

func (s *command) Usage() string {
	return Usage
}

func (s *command) Name() string {
	return Name
}

func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)
	yarnOnly := cmd.Bool("y", false, "Yarn only")
	goOnly := cmd.Bool("g", false, "Go only")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	} else if yarnOnly != goOnly {
		s.yarnOnly = *yarnOnly
		s.goOnly = *goOnly
	}

	return nil
}

func (s *command) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if _, err := workspace.ParseJSAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing yarn workspace absolute paths: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing module absolute paths: %w", err)
	} else if err := s.runTests(cwd, modPaths); err != nil {
		return fmt.Errorf("running tests: %w", err)
	} else {
		return nil
	}
}

func (s *command) runTests(cwd string, modPaths []string) error {
	if !s.goOnly {
		if err := yarn.TestWorkspace(cwd, s.env); err != nil {
			return fmt.Errorf("testing yarn workspace: %w", err)
		}
	}

	if !s.yarnOnly {
		if err := workspace.TestWorkspace(cwd, modPaths, s.env); err != nil {
			return fmt.Errorf("testing go workspace: %w", err)
		}
	}

	return nil
}
