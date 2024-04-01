package deps

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

const (
	Name  = "deps"
	Usage = "Ensure workspace dependencies are up to date"
)

type command struct {
	env environment.Environment
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

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

func (s *command) Run() error {
	fmt.Println("Show Report")
	return nil
}
