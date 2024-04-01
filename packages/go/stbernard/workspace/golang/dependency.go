package golang

import (
	"fmt"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// InstallGolangCiLint runs go install for the currently supported golangci-lint version
func InstallGolangCiLint(path string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"install", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2"}
	)

	if err := cmdrunner.Run(command, args, path, env); err != nil {
		return fmt.Errorf("golangci-lint install: %w", err)
	}

	return nil
}
