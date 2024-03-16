package cmdrunner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/specterops/bloodhound/log"
)

// Run runs a command with args
// If debug log level is set globally, command output will be combined and sent to os.Stderr
func Run(command string, args []string) error {
	var (
		cmdstr = command + strings.Join(args, " ")
		cmd    = genCmd(command, args)
	)

	log.Infof("Running %s", cmdstr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmdstr, err)
	} else {
		log.Infof("Finished %s", cmdstr)
		return nil
	}
}

// RunWithEnv runs a command with args and environment variables set
// If debug log level is set globally, command output will be combined and sent to os.Stderr
func RunWithEnv(command string, args []string, env []string) error {
	var (
		cmdstr = command + strings.Join(args, " ")
		cmd    = genCmd(command, args)
	)

	cmd.Env = env

	log.Infof("Running %s", cmdstr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmdstr, err)
	} else {
		log.Infof("Finished %s", cmdstr)
		return nil
	}
}

// RunAtPathWithEnv runs a command with ars and environment variables set at a specified path
// If debug log level is set globally, command output will be combined and sent to os.Stderr
func RunAtPathWithEnv(command string, args []string, path string, env []string) error {
	var (
		cmdstr = command + strings.Join(args, " ")
		cmd    = genCmd(command, args)
	)

	cmd.Env = env
	cmd.Dir = path

	log.Infof("Running %s for %s", cmdstr, path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmdstr, err)
	} else {
		log.Infof("Finished %s for %s", cmdstr, path)
		return nil
	}
}

// genCmd is used to generate a command with combined output sent to os.Stderr if debug global log level is set
func genCmd(command string, args []string) *exec.Cmd {
	cmd := exec.Command(command, args...)

	if log.GlobalAccepts(log.LevelDebug) {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	return cmd
}
