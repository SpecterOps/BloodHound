package envdump

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	Name  = "envdump"
	Usage = "Dump your environment variables"
)

type Config struct {
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
	log.Printf("Environment:\n\n")
	for _, env := range s.config.Environment {
		envTuple := strings.SplitN(env, "=", 2)
		log.Printf("%s: %s\n", envTuple[0], envTuple[1])
	}
	log.Printf("\n")

	return nil
}

func Create(config Config) (command, error) {
	envdumpCmd := flag.NewFlagSet(Name, flag.ExitOnError)

	envdumpCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n", Usage, filepath.Base(os.Args[0]), Name)
	}

	if err := envdumpCmd.Parse(os.Args[2:]); err != nil {
		envdumpCmd.Usage()
		return command{}, fmt.Errorf("failed to parse %s command: %w", Name, err)
	} else {
		return command{config: config}, nil
	}
}
