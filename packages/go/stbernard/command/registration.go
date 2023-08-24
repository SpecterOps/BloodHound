package command

import (
	"github.com/specterops/bloodhound/packages/go/stbernard/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/modsync"
)

// Command enum represents our subcommands
type Command int

const (
	InvalidCommand Command = iota - 1
	ModSync
	EnvDump
)

// String implements Stringer for Command enum
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

func Commands() []Command {
	return []Command{ModSync, EnvDump}
}

func CommandsUsage() []string {
	var usage = make([]string, len(Commands()))

	usage[ModSync] = modsync.Usage
	usage[EnvDump] = envdump.Usage

	return usage
}
