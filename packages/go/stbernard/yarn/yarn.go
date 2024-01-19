package yarn

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func InstallWorkspaceDeps(jsPaths []string, env []string) error {
	for _, path := range jsPaths {
		if err := yarnInstall(path, env); err != nil {
			return fmt.Errorf("failed to run yarn install at %v: %w", path, err)
		}
	}

	return nil
}

func yarnInstall(path string, env []string) error {
	cmd := exec.Command("yarn", "install")
	cmd.Env = env
	cmd.Dir = path
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	log.Printf("Running yarn install for %v\n", path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yarn install: %w", err)
	} else {
		log.Printf("Finished yarn install for %v\n", path)
		return nil
	}
}
