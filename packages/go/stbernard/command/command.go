package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/modsync"
)

type Commander interface {
	Name() string
	Usage() string
	Run() error
}

var NoCmdErr = errors.New("no command specified")
var InvalidCmdErr = errors.New("invalid command specified")
var FailedCreateCmdErr = errors.New("failed to create command")

func ParseCLI() (Commander, error) {
	// Generate a nice usage message
	flag.Usage = globalUsage

	// Default usage if no arguments provided
	if len(os.Args) < 2 {
		flag.Usage()
		return nil, NoCmdErr
	}

	switch os.Args[1] {
	case ModSync.String():
		config := modsync.Config{Environment: environment()}
		if cmd, err := modsync.CreateModSyncCommand(config); err != nil {
			return nil, fmt.Errorf("%w: %w", FailedCreateCmdErr, err)
		} else {
			return cmd, nil
		}

	case EnvDump.String():
		config := envdump.Config{Environment: environment()}
		if cmd, err := envdump.CreateEnvDumpCommand(config); err != nil {
			return nil, fmt.Errorf("%w: %w", FailedCreateCmdErr, err)
		} else {
			return cmd, nil
		}

	default:
		flag.Parse()
		flag.Usage()
		return nil, InvalidCmdErr
	}
}

func globalUsage() {
	var longestCmdLen int

	w := flag.CommandLine.Output()
	fmt.Fprint(w, "A BloodHound Swiss Army Knife\n\nUsage:  stbernard COMMAND\n\nCommands:\n")

	for _, cmd := range Commands() {
		if len(cmd.String()) > longestCmdLen {
			longestCmdLen = len(cmd.String())
		}
	}

	for cmd, usage := range CommandsUsage() {
		cmdStr := Command(cmd).String()
		padding := strings.Repeat(" ", longestCmdLen-len(cmdStr))
		fmt.Fprintf(w, "  %s%s    %s\n", cmdStr, padding, usage)
	}
}

// environment is used to add default env vars as needed to the existing environment variables
func environment() []string {
	var envMap = make(map[string]string)

	for _, env := range os.Environ() {
		envTuple := strings.SplitN(env, "=", 2)
		envMap[envTuple[0]] = envTuple[1]
	}

	// Make any changes here
	envMap["FOO"] = "foo" // For illustrative purposes only

	var envSlice = make([]string, 0, len(envMap))
	for key, val := range envMap {
		envSlice = append(envSlice, strings.Join([]string{key, val}, "="))
	}

	return envSlice
}
