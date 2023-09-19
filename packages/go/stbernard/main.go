package main

import (
	"errors"
	"log"

	"github.com/specterops/bloodhound/packages/go/stbernard/command"
)

func main() {
	if cmd, err := command.ParseCLI(); err != nil {
		if errors.Is(err, command.NoCmdErr) {
			log.Fatal("No command specified")
		} else {
			log.Fatalf("Error while parsing command: %v", err)
		}
	} else if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run command %s: %v", cmd.Name(), err)
	} else {
		log.Printf("%s completed successfully", cmd.Name())
	}
}
