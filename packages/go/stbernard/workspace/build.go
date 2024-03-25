package workspace

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

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
				go func(p GoPackage) {
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
