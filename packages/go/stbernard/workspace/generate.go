package workspace

import (
	"errors"
	"fmt"
	"sync"
)

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
