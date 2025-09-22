package alluregen

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	alluregen "github.com/specterops/bloodhound/packages/go/stbernard/command/alluregen/internal"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

const (
	Name  = "alluregen"
	Usage = "Generate Allure Result for Go tests"
)

type command struct {
	env environment.Environment
}

// Create a new instance of license command within the current environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of the command
func (s *command) Usage() string {
	return Usage
}

// Name of the command
func (s *command) Name() string {
	return Name
}

func (c *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)
	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

func (c *command) Run() error {
	if err := alluregen.Run(); err != nil {
		return fmt.Errorf("Running alluregen cmd: %w", err)
	}
	return nil
}
