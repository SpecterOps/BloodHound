package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func TestWorkspace(cwd string, modPaths []string, env environment.Environment) error {
	for _, modPath := range modPaths {
		var (
			coverFile = filepath.Join(cwd, ".cache", filepath.Base(modPath)+".coverage")
			command   = "go"
			args      = []string{"test", "-coverprofile", coverFile, "./..."}
		)

		if err := cmdrunner.Run(command, args, modPath, env); err != nil {
			return fmt.Errorf("go test at %v: %w", modPath, err)
		}
	}

	return nil
}
