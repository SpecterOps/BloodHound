package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
)

func main() {
	fmt.Println("🧪 Proving Compile-Time Provider Switching")
	fmt.Println("==========================================")

	// Create service with temp file (empty path uses provider defaults)
	config := dogtags.Config{FilePath: ""}
	service, err := dogtags.NewService(config)
	if err != nil {
		fmt.Printf("❌ Failed to create service: %v\n", err)
		return
	}

	ctx := context.Background()
	allFlags := service.GetAllFlags(ctx)

	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("  Number of flags loaded: %d\n", len(allFlags))

	fmt.Printf("\n  Flags:\n")

	// Sort keys for consistent output
	var keys []string
	for key := range allFlags {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("    %-20s: %v\n", key, allFlags[dogtags.FlagKey(key)])
	}

	fmt.Printf("\n✅ Provider compiled in via build tags\n")

	fmt.Printf("\n🔧 To test:\n")
	fmt.Printf("  YAML Provider: go run prove_switching_test.go\n")
	fmt.Printf("  No-Op Provider: go run -tags=noop prove_switching_test.go\n")
}
