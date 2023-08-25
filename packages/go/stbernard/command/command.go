package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
)

// Commander is an interface for commands, allowing commands to implement the minimum
// set of requirements to observe and run the command from above. It is used as a return
// type to allow passing a usable command to the caller after parsing and creating
// the command implementation
type Commander interface {
	Name() string
	Usage() string
	Run() error
}

var NoCmdErr = errors.New("no command specified")
var InvalidCmdErr = errors.New("invalid command specified")
var FailedCreateCmdErr = errors.New("failed to create command")

// ParseCLI parses for a subcommand as the first argument to the calling binary,
// and initializes the command (if it exists). It also provides the default usage
// statement.
//
// It does not support flags of its own, each subcommand is responsible for parsing
// their flags.
func ParseCLI() (Commander, error) {
	// Generate a nice usage message
	flag.Usage = usage

	// Default usage if no arguments provided
	if len(os.Args) < 2 {
		flag.Usage()
		return nil, NoCmdErr
	}

	switch os.Args[1] {
	case ModSync.String():
		config := modsync.Config{Environment: environment()}
		if cmd, err := modsync.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", FailedCreateCmdErr, err)
		} else {
			return cmd, nil
		}

	case EnvDump.String():
		config := envdump.Config{Environment: environment()}
		if cmd, err := envdump.Create(config); err != nil {
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

// usage creates a pretty usage message for our main command
func usage() {
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
