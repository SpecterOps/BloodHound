package golang

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
	"golang.org/x/mod/modfile"
)

const (
	CoverageManifest = "manifest.json"
	CoverageExt      = ".coverage"
	CombinedCoverage = "combined" + CoverageExt
)

var (
	DefaultCoveragePath          = filepath.Join("tmp", "coverage")
	DefaultIntegrationConfigPath = filepath.Join("local-harnesses", "integration.config.json")
)

func TestWorkspace(cwd string, modPaths []string, profileDir string, env environment.Environment, integration bool) error {
	var (
		manifest = make(map[string]string, len(modPaths))
		command  = "go"
		args     = []string{"test"}
	)

	if integration {
		if integrationConfigPath, ok := env["INTEGRATION_CONFIG_PATH"]; !ok || integrationConfigPath == "" {
			env["INTEGRATION_CONFIG_PATH"] = filepath.Join(cwd, DefaultIntegrationConfigPath)
		} else if !filepath.IsAbs(integrationConfigPath) {
			env["INTEGRATION_CONFIG_PATH"] = filepath.Join(cwd, integrationConfigPath)
		}

		args = append(args, []string{"-p", "1", "-tags", "integration serial_integration"}...)
	}

	for _, modPath := range modPaths {
		modName, err := GetModuleName(modPath)
		if err != nil {
			return err
		}

		fileUUID, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("create uuid for coverfile: %w", err)
		}

		coverFile := filepath.Join(profileDir, fileUUID.String()+".coverage")
		manifest[modName] = coverFile
		testArgs := slicesext.Concat(args, []string{"-coverprofile", coverFile, "./..."})

		if err := cmdrunner.Run(command, testArgs, modPath, env); err != nil {
			return fmt.Errorf("go test at %v: %w", modPath, err)
		}
	}

	if manifestFile, err := os.Create(filepath.Join(profileDir, CoverageManifest)); err != nil {
		return fmt.Errorf("create manifest file: %w", err)
	} else if marshaledManifest, err := json.Marshal(manifest); err != nil {
		manifestFile.Close()
		return fmt.Errorf("marshal manifest: %w", err)
	} else if _, err := manifestFile.Write(marshaledManifest); err != nil {
		manifestFile.Close()
		return fmt.Errorf("writing manifest to file: %w", err)
	} else if err := manifestFile.Close(); err != nil {
		return fmt.Errorf("closing manifest file: %w", err)
	} else {
		return nil
	}
}

func GetModuleName(modPath string) (string, error) {
	if modFile, err := os.ReadFile(filepath.Join(modPath, "go.mod")); err != nil {
		return "", fmt.Errorf("reading go.mod file for %s: %w", modPath, err)
	} else {
		return modfile.ModulePath(modFile), nil
	}
}
