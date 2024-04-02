package tester

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/yarn"
)

const (
	Name  = "test"
	Usage = "Run tests for entire workspace"
)

type command struct {
	env         environment.Environment
	yarnOnly    bool
	goOnly      bool
	integration bool
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
	integration := cmd.Bool("i", false, "Include integration tests")

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

	s.integration = *integration

	return nil
}

func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if _, err := yarn.ParseWorkspace(paths.Root); err != nil {
		return fmt.Errorf("parsing yarn workspace absolute paths: %w", err)
	} else if err := s.runTests(paths.Root, paths.Coverage, paths.GoModules); err != nil {
		return fmt.Errorf("running tests: %w", err)
	} else {
		return nil
	}
}

func (s *command) runTests(cwd string, coverPath string, modPaths []string) error {
	if !s.goOnly {
		if err := yarn.TestWorkspace(cwd, s.env); err != nil {
			return fmt.Errorf("testing yarn workspace: %w", err)
		}
	}

	if !s.yarnOnly {
		log.Infof("Checking coverage directory")
		if err := os.MkdirAll(coverPath, os.ModeDir+fs.ModePerm); err != nil {
			return fmt.Errorf("making coverage directory: %w", err)
		} else if dirList, err := os.ReadDir(coverPath); err != nil {
			return fmt.Errorf("listing coverage directory: %w", err)
		} else {
			for _, entry := range dirList {
				if filepath.Ext(entry.Name()) == golang.CoverageExt {
					log.Debugf("Removing %s", filepath.Join(coverPath, entry.Name()))
					if err := os.Remove(filepath.Join(coverPath, entry.Name())); err != nil {
						return fmt.Errorf("removing %s: %w", filepath.Join(coverPath, entry.Name()), err)
					}
				}
			}
		}

		if err := golang.TestWorkspace(cwd, modPaths, coverPath, s.env, s.integration); err != nil {
			return fmt.Errorf("testing go workspace: %w", err)
		}
	}

	return nil
}
