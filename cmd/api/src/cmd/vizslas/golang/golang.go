package golang

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/cmd/vizslas/cve"
	"github.com/specterops/bloodhound/src/cmd/vizslas/golang/vulndb"
	"github.com/specterops/bloodhound/src/cmd/vizslas/ingest"
	"github.com/specterops/bloodhound/src/cmd/vizslas/ioutils"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

type CommandError struct {
	Message string
	Cause   error
	Raw     []byte
}

func (s *CommandError) Error() string {
	return fmt.Sprintf("%s: %s", s.Message, s.Raw)
}

func jsonDecodeSlice[T any, S []T](reader io.Reader) (S, error) {
	var (
		slice   S
		decoder = json.NewDecoder(reader)
	)

	for decoder.More() {
		var value T

		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}

		slice = append(slice, value)
	}

	return slice, nil
}

func jsonDecodeElement[T any](reader io.Reader) (T, error) {
	var (
		value   T
		decoder = json.NewDecoder(reader)
	)

	return value, decoder.Decode(&value)
}

func ListModule(moduleDir string) (*Module, error) {
	var module *Module

	return module, PushDirectory(moduleDir, func() error {
		if output, err := Exec([]string{"go", "list", "-json", "-m"}); err != nil {
			return AsCommandError("ListModule", output, err)
		} else {
			module, err = jsonDecodeElement[*Module](bytes.NewBuffer(output))
			return err
		}
	})
}

func ListWorkspaceModules(workspaceDir string) ([]*Module, error) {
	var modules []*Module

	// List all workspace modules first
	if err := PushDirectory(workspaceDir, func() error {
		if output, err := Exec([]string{"go", "list", "-m", "-json", "all"}); err != nil {
			return err
		} else {
			modules, err = jsonDecodeSlice[*Module](bytes.NewBuffer(output))
			return err
		}
	}); err != nil {
		return nil, err
	}

	// Collect the GOROOT stdlib module last
	if rootModule, err := ListRootModule(); err != nil {
		return nil, err
	} else {
		modules = append(modules, rootModule)
	}

	return modules, nil
}

func ListPackages(module *Module) ([]*Package, error) {
	var pkgs []*Package

	return pkgs, PushDirectory(module.Dir, func() error {
		if output, err := Exec([]string{"go", "list", "-json", "./..."}); err != nil {
			return AsCommandError("ListPackages", output, err)
		} else if pkgs, err = jsonDecodeSlice[*Package](bytes.NewBuffer(output)); err != nil {
			return err
		}

		// Tie the packages to the actual module pointer
		for _, pkg := range pkgs {
			pkg.Module = module
		}

		return nil
	})
}

func AsCommandError(message string, output []byte, err error) error {
	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		return &CommandError{
			Message: message,
			Cause:   err,
			Raw:     exitErr.Stderr,
		}
	}

	return &CommandError{
		Message: message,
		Cause:   err,
		Raw:     output,
	}
}

func GetGoRoot() (string, error) {
	if output, err := Exec([]string{"go", "env", "GOROOT"}); err != nil {
		return "", AsCommandError("go env GOROOT", output, err)
	} else {
		return strings.TrimSpace(string(output)), nil
	}
}

func ListRootModule() (*Module, error) {
	if goRootPath, err := GetGoRoot(); err != nil {
		return nil, err
	} else {
		goSrcPath := path.Join(goRootPath, "src")

		return ListModule(goSrcPath)
	}
}

func ListModules(path string) ([]*Module, error) {
	var modules []*Module

	return modules, PushDirectory(path, func() error {
		if output, err := Exec([]string{"go", "list", "-m", "-json", "all"}); err != nil {
			return err
		} else {
			decoder := json.NewDecoder(bytes.NewBuffer(output))

			for decoder.More() {
				var module Module

				if err := decoder.Decode(&module); err != nil {
					return err
				}

				modules = append(modules, &module)
			}

			return nil
		}
	})
}

var (
	cveFilter = ioutils.FileFilter{
		MustContainOneOf: []string{"2023", "2024"},
		AcceptsExtension: func(filePath string, fileInfo fs.FileInfo) bool {
			baseName := path.Base(filePath)
			return !fileInfo.IsDir() && strings.HasPrefix(baseName, "CVE") && path.Ext(baseName) == ".json"
		},
	}

	vlundbFilter = ioutils.FileFilter{
		MustContainOneOf: []string{"ID"},
		AcceptsExtension: func(filePath string, fileInfo fs.FileInfo) bool {
			return !fileInfo.IsDir() &&
				path.Base(filePath) != "index.json" &&
				path.Ext(filePath) == ".json"
		},
	}
)

func ListWorkspace(workspaceDir string) (*Workspace, error) {
	workspace := NewWorkspace()

	if modules, err := ListWorkspaceModules(workspaceDir); err != nil {
		return nil, err
	} else {
		for _, module := range modules {
			workspace.AddModule(module)

			if modulePkgs, err := ListPackages(module); err != nil {
				log.Warnf("Unable to list packages for module %s(%s): %v", module.Dir, module.Path, err)
			} else {
				for _, modulePkg := range modulePkgs {
					workspace.AddPackage(modulePkg)
				}
			}
		}
	}

	return workspace, nil
}

func DownloadGoVulnDatabaseArchive(path string) (int64, error) {
	if request, err := http.NewRequest(http.MethodGet, "https://vuln.go.dev/vulndb.zip", nil); err != nil {
		return 0, err
	} else {
		request.Header.Add("User-Agent", "bh_govuln-0.0.0")

		if fout, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664); err != nil {
			return 0, err
		} else {
			defer fout.Close()

			if resp, err := http.DefaultClient.Do(request); err != nil {
				return 0, err
			} else {
				defer resp.Body.Close()
				return io.Copy(fout, resp.Body)
			}
		}
	}
}

func ReadVulnDBArchive(ctx context.Context, archivePath string, filter ioutils.FileFilter) (*vulndb.Database, error) {
	if fileIn, err := os.Open(archivePath); err != nil {
		return nil, err
	} else {
		defer fileIn.Close()

		if zipFileInfo, err := fileIn.Stat(); err != nil {
			return nil, err
		} else if zipReader, err := zip.NewReader(fileIn, zipFileInfo.Size()); err != nil {
			return nil, err
		} else if database, err := vulndb.New(); err != nil {
			return nil, err
		} else {
			for _, file := range zipReader.File {
				// Check the context for early exit
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				// Check to see if this file is accepted
				if !filter.Accepts(file.Name, file.FileInfo()) {
					continue
				}

				// Read and marshall the CVE JSON file
				if entry, err := ioutils.JSONDecodeZipFile[vulndb.Entry](zipReader, file.Name); err != nil {
					log.Errorf("Zip file reader error: %v", err)
				} else if err := database.Add(entry); err != nil {
					log.Errorf("Failed to add entry to vlundb: %v", err)
				}
			}

			log.Infof("Loaded %d entries", len(database.Entries))

			return database, nil
		}
	}
}

func AnalyzeGoWorkspace(ctx context.Context, workspaceDir, cveDBPath, govulnDBPath string) (ingest.Payload, error) {
	if cveDB, err := cve.ReadArchive(ctx, cveDBPath, cveFilter); err != nil {
		return ingest.Payload{}, fmt.Errorf("failed reading CVE database archive %s: %w", err)
	} else if govulnDB, err := ReadVulnDBArchive(ctx, govulnDBPath, vlundbFilter); err != nil {
		return ingest.Payload{}, fmt.Errorf("failed reading govlun database archive %s: %w", govulnDBPath, err)
	} else if workspace, err := ListWorkspace(workspaceDir); err != nil {
		return ingest.Payload{}, fmt.Errorf("failed reading go workspace %s: %w", workspaceDir, err)
	} else {
		workspace.cveDB = cveDB
		workspace.govlunDB = govulnDB

		return ToIngestPayload(workspace), nil
	}
}
