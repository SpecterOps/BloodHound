package command

import (
	"github.com/specterops/bloodhound/packages/go/stbernard/foo"
	"github.com/specterops/bloodhound/packages/go/stbernard/modsync"
)

// Command enum represents our subcommands
type Command int

const (
	InvalidCommand Command = iota - 1
	ModSync
	Foo
)

// String implements Stringer for Command enum
func (s Command) String() string {
	switch s {
	case ModSync:
		return modsync.Name
	case Foo:
		return foo.Name
	default:
		return "invalid command"
	}
}

func Commands() []Command {
	return []Command{ModSync, Foo}
}

func CommandsUsage() []string {
	var usage = make([]string, len(Commands()))

	usage[ModSync] = modsync.Usage
	usage[Foo] = foo.Usage

	return usage
}
