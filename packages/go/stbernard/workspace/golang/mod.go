package golang

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("reading go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("parsing go.work file: %w", err)
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

// moduleListPackages runs go list for the given module and returns the list of packages in that module
func moduleListPackages(modPath string) ([]GoPackage, error) {
	var (
		packages = make([]GoPackage, 0)
	)

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = modPath
	if out, err := cmd.StdoutPipe(); err != nil {
		return packages, fmt.Errorf("creating stdout pipe for module %s: %w", modPath, err)
	} else if err := cmd.Start(); err != nil {
		return packages, fmt.Errorf("listing packages for module %s: %w", modPath, err)
	} else {
		decoder := json.NewDecoder(out)
		for {
			var p GoPackage
			if err := decoder.Decode(&p); err == io.EOF {
				break
			} else if err != nil {
				return packages, fmt.Errorf("decoding package in module %s: %w", modPath, err)
			}
			packages = append(packages, p)
		}
		cmd.Wait()
		return packages, nil
	}
}
