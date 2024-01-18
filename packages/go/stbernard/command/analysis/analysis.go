package analysis

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "analysis"
	Usage = "Run static analyzers"
)

type Config struct {
	Environment []string
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
		return fmt.Errorf("could not find workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse module absolute paths: %w", err)
	} else if jsPaths, err := workspace.ParseJsAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse JS absolute paths: %w", err)
	} else if result, err := analyzers.Run(cwd, modPaths, jsPaths, s.config.Environment); errors.Is(err, analyzers.ErrSeverityExit) {
		fmt.Println(result)
		return err
	} else if err != nil {
		return fmt.Errorf("analyzers could not run completely: %w", err)
	} else {
		fmt.Println(result)
		return nil
	}
}

func Create(config Config) (command, error) {
	analysisCmd := flag.NewFlagSet(Name, flag.ExitOnError)

	analysisCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		analysisCmd.PrintDefaults()
	}

	if err := analysisCmd.Parse(os.Args[2:]); err != nil {
		analysisCmd.Usage()
		return command{}, fmt.Errorf("failed to parse analysis command: %w", err)
	} else {
		return command{config: config}, nil
	}
}
