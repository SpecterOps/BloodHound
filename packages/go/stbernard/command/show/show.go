package show

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
	Name  = "show"
	Usage = "Show current project info"
)

type command struct {
	env environment.Environment
}

type repository struct {
	path    string
	sha     string
	version semver.Version
	clean   bool
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
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace paths: %w", err)
	} else if rootRepo, err := s.repositoryCheck(paths.Root); err != nil {
		return fmt.Errorf("repository check: %w", err)
	} else if submodules, err := s.submodulesCheck(paths.Submodules); err != nil {
		return fmt.Errorf("submodule check: %w", err)
	} else {
		repos := []repository{rootRepo}
		repos = append(repos, submodules...)
		for _, repo := range repos {
			fmt.Printf("Repository Report For %s\n", repo.path)
			fmt.Printf("Current HEAD: %s\n", repo.sha)
			fmt.Printf("Detected versions: %s\n", repo.version)
			if !repo.clean {
				fmt.Println("CHANGES DETECTED")
				return fmt.Errorf("changes detected in git repository")
			} else {
				fmt.Printf("Repository Clean\n\n")
			}
		}

		return nil
	}
}

func (s *command) submodulesCheck(paths []string) ([]repository, error) {
	var submodules = make([]repository, 0, len(paths))

	for _, path := range paths {
		if repo, err := s.repositoryCheck(path); err != nil {
			return submodules, fmt.Errorf("checking repository for submodule %s: %w", path, err)
		} else {
			submodules = append(submodules, repo)
		}
	}

	return submodules, nil
}

func (s *command) repositoryCheck(cwd string) (repository, error) {
	var repo repository

	if sha, err := git.FetchCurrentCommitSHA(cwd, s.env); err != nil {
		return repo, fmt.Errorf("fetching current commit sha: %w", err)
	} else if version, err := git.ParseLatestVersionFromTags(cwd, s.env); err != nil {
		return repo, fmt.Errorf("parsing version: %w", err)
	} else if clean, err := git.CheckClean(cwd, s.env); err != nil {
		return repo, fmt.Errorf("checking repository clean: %w", err)
	} else {
		repo.path = cwd
		repo.sha = sha
		repo.version = version
		repo.clean = clean

		return repo, nil
	}
}
