package coverfiles

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

var (
	// ErrNoProfiles tells the caller that no profiles were provided
	ErrNoProfiles = errors.New("no profiles")
)

func WriteProfile(w io.Writer, profiles []*cover.Profile) error {
	if len(profiles) == 0 {
		return ErrNoProfiles
	}

	if _, err := fmt.Fprintf(w, "mode: %s\n", profiles[0].Mode); err != nil {
		return err
	}

	for _, p := range profiles {
		for _, b := range p.Blocks {
			if _, err := fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count); err != nil {
				return err
			}
		}
	}

	return nil
}
