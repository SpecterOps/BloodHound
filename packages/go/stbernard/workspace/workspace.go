package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
)

type Package struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	Import string `json:"importpath"`
}

// FindRoot will attempt to crawl up the path until it finds a go.work file
func FindRoot() (string, error) {
	if cwd, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	} else {
		var found bool

		for !found {
			found, err = workFileExists(cwd)
			if err != nil {
				return cwd, fmt.Errorf("error while trying to find go.work file: %w", err)
			}

			if found {
				break
			}

			prevCwd := cwd

			// Go up a directory before retrying
			cwd = filepath.Dir(cwd)

			if cwd == prevCwd {
				return cwd, errors.New("found root path without finding go.work file")
			}
		}

		return cwd, nil
	}
}

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
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

// DownloadModules runs go mod download for all module paths passed
func DownloadModules(modPaths []string) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			wg.Done()
			cmd := exec.Command("go", "mod", "download")
			cmd.Dir = modPath
			if err := cmd.Run(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failure when running go mod download in %s: %w", modPath, err))
				mu.Unlock()
			}
		}(modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// WorkspaceGenerate runs go generate ./... for all module paths passed
func WorkspaceGenerate(modPaths []string) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			defer wg.Done()
			if err := moduleGenerate(modPath); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failure running code generation for module %s: %w", modPath, err))
				mu.Unlock()
			}
		}(modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// SyncWorkspace runs go work sync in the given directory
func SyncWorkspace(cwd string) error {
	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed running go work sync: %w", err)
	} else {
		return nil
	}
}

// BuildMainPackages builds all main packages for a list of module paths
func BuildMainPackages(workRoot string, modPaths []string) error {
	var (
		errs     []error
		wg       sync.WaitGroup
		mu       sync.Mutex
		buildDir = filepath.Join(workRoot, "dist") + string(filepath.Separator)
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(buildDir, modPath string) {
			defer wg.Done()
			if err := buildModuleMainPackages(buildDir, modPath); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to build main package: %w", err))
				mu.Unlock()
			}
		}(buildDir, modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// moduleListPackages runs go list for the given module and returns the list of packages in that module
func moduleListPackages(modPath string) ([]Package, error) {
	var (
		packages = make([]Package, 0)
	)

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = modPath
	if out, err := cmd.StdoutPipe(); err != nil {
		return packages, fmt.Errorf("failed to create stdout pipe for module %s: %w", modPath, err)
	} else if err := cmd.Start(); err != nil {
		return packages, fmt.Errorf("failed to list packages for module %s: %w", modPath, err)
	} else {
		decoder := json.NewDecoder(out)
		for {
			var p Package
			if err := decoder.Decode(&p); err == io.EOF {
				break
			} else if err != nil {
				return packages, fmt.Errorf("failed to decode package in module %s: %w", modPath, err)
			}
			packages = append(packages, p)
		}
		cmd.Wait()
		return packages, nil
	}
}

// buildModuleMainPackages runs go build for all main packages in a given module
func buildModuleMainPackages(buildDir string, modPath string) error {
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex
	)

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("failed to list module packages: %w", err)
	} else {
		for _, p := range packages {
			if p.Name == "main" && !strings.Contains(p.Dir, "plugin") {
				wg.Add(1)
				go func(p Package) {
					defer wg.Done()
					cmd := exec.Command("go", "build", "-o", buildDir)
					cmd.Dir = p.Dir
					if err := cmd.Run(); err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("failed running go build for package %s: %w", p.Import, err))
						mu.Unlock()
					} else {
						slog.Info("Built package", "package", p.Import, "dir", p.Dir)
					}
				}(p)
			}
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}

// workFileExists checks if a go.work file exists in the given directory
func workFileExists(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not stat go.work file: %w", err)
	} else {
		return true, nil
	}
}

// moduleGenerate runs go generate in each package of the given module
func moduleGenerate(modPath string) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("could not list packages for module %s: %w", modPath, err)
	} else {
		for _, pkg := range packages {
			wg.Add(1)
			go func(pkg Package) {
				defer wg.Done()
				cmd := exec.Command("go", "generate", pkg.Dir)
				cmd.Dir = modPath
				slog.Info("Generating code for package", "package", pkg.Name, "path", pkg.Dir)
				if err := cmd.Run(); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to generate code for package %s: %w", pkg, err))
					mu.Unlock()
				}
			}(pkg)
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
