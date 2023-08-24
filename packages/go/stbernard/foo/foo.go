package foo

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	Name  = "foo"
	Usage = "Foo does foo things"
)

type flags struct {
	verbose bool
}

type Config struct {
	flags       flags
	Environment []string
}

type command struct {
	config Config
}

func (s command) Name() string {
	return Name
}

func (s command) Usage() string {
	return Usage
}

func (s command) Run() error {
	log.Printf("Flags: %+v\n", s.config.flags)
	log.Printf("Environment: %#v\n", s.config.Environment)

	return nil
}

func CreateFooCommand(config Config) (command, error) {
	fooCmd := flag.NewFlagSet(Name, flag.ExitOnError)
	fooCmd.BoolVar(&config.flags.verbose, "v", false, "Print verbose logs")

	fooCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		fooCmd.PrintDefaults()
	}

	if err := fooCmd.Parse(os.Args[2:]); err != nil {
		fooCmd.Usage()
		return command{}, fmt.Errorf("failed to parse %s command: %w", Name, err)
	} else {
		return command{config: config}, nil
	}
}
