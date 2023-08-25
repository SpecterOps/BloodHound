package command

import (
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
)

// Command enum represents our subcommands
type Command int

const (
	InvalidCommand Command = iota - 1
	ModSync
	EnvDump
)

// String implements Stringer for the Command enum
func (s Command) String() string {
	switch s {
	case ModSync:
		return modsync.Name
	case EnvDump:
		return envdump.Name
	default:
		return "invalid command"
	}
}

// Commands returns our valid set of Command options
func Commands() []Command {
	return []Command{ModSync, EnvDump}
}

// Commands usage returns a slice of Command usage statements indexed by their enum
func CommandsUsage() []string {
	var usage = make([]string, len(Commands()))

	usage[ModSync] = modsync.Usage
	usage[EnvDump] = envdump.Usage

	return usage
}
