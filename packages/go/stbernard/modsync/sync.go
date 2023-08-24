package modsync

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// findRoot will attempt to crawl up the path until it finds a go.work file
func findRoot() (string, error) {
	if cwd, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	} else {
		var found bool

		for !found {
			found, err = findWorkFile(cwd)
			if err != nil {
				return cwd, fmt.Errorf("error while trying to find go.work file: %w", err)
			}

			if found {
				break
			}

			// Go up a directory before retrying
			cwd = filepath.Dir(cwd)
		}
		return cwd, nil
	}
}

func findWorkFile(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not stat go.work file: %w", err)
	} else {
		return true, nil
	}
}

func parseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("could not read go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("could not parse go.work file: %w", err)
	} else {
		var (
			modulePaths = make([]string, 0, len(workfile.Use))
			workDir     = filepath.Dir(workfilePath)
		)

		for _, use := range workfile.Use {
			modulePaths = append(modulePaths, filepath.Join(workDir, use.Path))
		}

		return modulePaths, nil
	}
}

func downloadMods(modPaths []string, env []string) error {
	var errs = make([]error, 0)

	for _, modPath := range modPaths {
		cmd := exec.Command("go", "mod", "download")
		cmd.Env = env
		cmd.Dir = modPath
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failure when running command: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to download all modules: %w", errors.Join(errs...))
	} else {
		return nil
	}
}

func syncWorkspace(cwd string, env []string) error {
	cmd := exec.Command("go", "work", "sync")
	cmd.Env = env
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed running go work sync: %w", err)
	} else {
		return nil
	}
}
