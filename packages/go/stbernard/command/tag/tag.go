package tag

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "tag"
	Usage = "Print current tag information for docker"
)

type command struct {
	env        environment.Environment
	primary    bool
	additional bool
}

// Create new instance of command to capture given environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of command
func (s *command) Usage() string {
	return Usage
}

// Name of command
func (s *command) Name() string {
	return Name
}

// Parse command flags
func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)

	cmd.BoolVar(&s.primary, "primary", false, "Print primary tag only")
	cmd.BoolVar(&s.additional, "additional", false, "Print additional tags")

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

// Run show command
func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace paths: %w", err)
	} else if sha, err := git.FetchCurrentCommitSHA(paths.Root, s.env); err != nil {
		return fmt.Errorf("fetching commit sha for tagging: %w", err)
	} else if version, err := git.ParseLatestVersionFromTags(paths.Root, s.env); err != nil {
		return fmt.Errorf("parsing version for tagging: %w", err)
	} else if err := s.printTags(sha, version); err != nil {
		return fmt.Errorf("printing tags: %w", err)
	} else {
		return nil
	}
}

func (s *command) printTags(sha string, version semver.Version) error {
	if s.primary {
		_, err := fmt.Fprintln(os.Stdout, version.String())
		return err
	} else if s.additional && version.Prerelease() != "" {
		_, err := fmt.Fprintf(os.Stdout, "candidate-%s candidate", sha)
		return err
	} else if s.additional && version.Prerelease() == "" {
		_, err := fmt.Fprintf(os.Stdout, "production-%s production", sha)
		return err
	} else {
		return fmt.Errorf("must provide a flag for either primary or additional")
	}
}
