package yarn

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/specterops/bloodhound/log"
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
	if log.GlobalAccepts(log.LevelDebug) {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	log.Infof("Running yarn install for %v", path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yarn install: %w", err)
	} else {
		log.Infof("Finished yarn install for %v", path)
		return nil
	}
}
