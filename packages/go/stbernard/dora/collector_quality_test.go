// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package dora

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// TestParseTagsToDeployments tests the full tag parsing pipeline with real semver tags
func TestParseTagsToDeployments(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "BloodHound",
		},
	}

	collector := NewGitHubCollector(config, env)
	now := time.Now()

	// Simulate a realistic release cycle with RCs and patches
	tags := []tagWithCommit{
		// v9.1.0 release cycle
		{Name: "v9.1.0-rc1", SHA: "sha1", Timestamp: now.Add(-10 * 24 * time.Hour)},
		{Name: "v9.1.0-rc2", SHA: "sha2", Timestamp: now.Add(-8 * 24 * time.Hour)},
		{Name: "v9.1.0", SHA: "sha3", Timestamp: now.Add(-7 * 24 * time.Hour)},
		{Name: "v9.1.1", SHA: "sha4", Timestamp: now.Add(-5 * 24 * time.Hour)},
		{Name: "v9.1.2", SHA: "sha5", Timestamp: now.Add(-3 * 24 * time.Hour)},

		// v9.2.0 release cycle
		{Name: "v9.2.0-rc1", SHA: "sha6", Timestamp: now.Add(-2 * 24 * time.Hour)},
		{Name: "v9.2.0", SHA: "sha7", Timestamp: now.Add(-1 * 24 * time.Hour)},

		// Non-semver tags (should be ignored)
		{Name: "latest", SHA: "sha8", Timestamp: now},
		{Name: "v9", SHA: "sha9", Timestamp: now},
	}

	startTime := now.Add(-15 * 24 * time.Hour)
	endTime := now

	deployments := collector.parseTagsToDeployments(tags, startTime, endTime)

	// Verify we got the right number of deployments (7 valid semver tags)
	if len(deployments) != 7 {
		t.Errorf("Expected 7 deployments, got %d", len(deployments))
	}

	// Find v9.1.0 and verify quality metrics
	var v910 *Deployment
	for i := range deployments {
		if deployments[i].Tag == "v9.1.0" {
			v910 = &deployments[i]
			break
		}
	}

	if v910 == nil {
		t.Fatal("v9.1.0 deployment not found")
		return
	}

	// v9.1.0 should have 2 RCs before it
	if v910.TotalRCs != 2 {
		t.Errorf("Expected v9.1.0 to have 2 RCs, got %d", v910.TotalRCs)
	}

	// v9.1.0 should have 2 patches after it (v9.1.1, v9.1.2)
	if v910.TotalPatches != 2 {
		t.Errorf("Expected v9.1.0 to have 2 patches, got %d", v910.TotalPatches)
	}

	// Verify v9.1.1 is marked as a patch
	var v911 *Deployment
	for i := range deployments {
		if deployments[i].Tag == "v9.1.1" {
			v911 = &deployments[i]
			break
		}
	}

	if v911 == nil {
		t.Fatal("v9.1.1 deployment not found")
		return
	}

	if !v911.IsPatch {
		t.Error("Expected v9.1.1 to be marked as patch")
	}

	if v911.PatchNumber != 1 {
		t.Errorf("Expected v9.1.1 to have patch number 1, got %d", v911.PatchNumber)
	}

	// Verify v9.1.0-rc1 is marked as RC
	var v910rc1 *Deployment
	for i := range deployments {
		if deployments[i].Tag == "v9.1.0-rc1" {
			v910rc1 = &deployments[i]
			break
		}
	}

	if v910rc1 == nil {
		t.Fatal("v9.1.0-rc1 deployment not found")
		return
	}

	if !v910rc1.IsRC {
		t.Error("Expected v9.1.0-rc1 to be marked as RC")
	}

	if v910rc1.RCNumber == nil {
		t.Error("Expected v9.1.0-rc1 to have RC number, got nil")
	} else if *v910rc1.RCNumber != 1 {
		t.Errorf("Expected v9.1.0-rc1 to have RC number 1, got %d", *v910rc1.RCNumber)
	}

	// Verify all deployments have correct IsProduction flag
	for _, d := range deployments {
		shouldBeProd := !d.IsRC
		if d.IsProduction != shouldBeProd {
			t.Errorf("Deployment %s: expected IsProduction=%v, got %v", d.Tag, shouldBeProd, d.IsProduction)
		}
	}

	// Verify HTML URLs are generated correctly
	expectedURL := "https://github.com/SpecterOps/BloodHound/releases/tag/v9.1.0"
	if v910.HTMLURL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, v910.HTMLURL)
	}
}

// TestParseTagsEdgeCases tests tag parsing with various edge cases
func TestParseTagsEdgeCases(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	config := &Config{
		GitHub: GitHubConfig{
			Owner: "test",
			Repo:  "test",
		},
	}

	collector := NewGitHubCollector(config, env)
	now := time.Now()

	tests := []struct {
		name          string
		tags          []tagWithCommit
		expectedCount int
		expectedRCs   map[string]int // version -> expected RC count
		expectedPatch map[string]int // version -> expected patch count
	}{
		{
			name: "multiple_rc_iterations",
			tags: []tagWithCommit{
				{Name: "v1.0.0-rc1", SHA: "sha1", Timestamp: now.Add(-10 * time.Hour)},
				{Name: "v1.0.0-rc2", SHA: "sha2", Timestamp: now.Add(-9 * time.Hour)},
				{Name: "v1.0.0-rc3", SHA: "sha3", Timestamp: now.Add(-8 * time.Hour)},
				{Name: "v1.0.0-rc4", SHA: "sha4", Timestamp: now.Add(-7 * time.Hour)},
				{Name: "v1.0.0-rc5", SHA: "sha5", Timestamp: now.Add(-6 * time.Hour)},
				{Name: "v1.0.0", SHA: "sha6", Timestamp: now.Add(-5 * time.Hour)},
			},
			expectedCount: 6,
			expectedRCs:   map[string]int{"1.0.0": 5},
			expectedPatch: map[string]int{"1.0.0": 0},
		},
		{
			name: "many_patches",
			tags: []tagWithCommit{
				{Name: "v2.0.0", SHA: "sha1", Timestamp: now.Add(-10 * time.Hour)},
				{Name: "v2.0.1", SHA: "sha2", Timestamp: now.Add(-9 * time.Hour)},
				{Name: "v2.0.2", SHA: "sha3", Timestamp: now.Add(-8 * time.Hour)},
				{Name: "v2.0.3", SHA: "sha4", Timestamp: now.Add(-7 * time.Hour)},
			},
			expectedCount: 4,
			expectedRCs:   map[string]int{"2.0.0": 0},
			expectedPatch: map[string]int{"2.0.0": 3},
		},
		{
			name: "mixed_invalid_tags",
			tags: []tagWithCommit{
				{Name: "v3.0.0", SHA: "sha1", Timestamp: now.Add(-5 * time.Hour)},
				{Name: "invalid-tag", SHA: "sha2", Timestamp: now.Add(-4 * time.Hour)},
				{Name: "v3.0.1", SHA: "sha3", Timestamp: now.Add(-3 * time.Hour)},
				{Name: "latest", SHA: "sha4", Timestamp: now.Add(-2 * time.Hour)},
				{Name: "v3", SHA: "sha5", Timestamp: now.Add(-1 * time.Hour)},
				{Name: "v3.0.2", SHA: "sha6", Timestamp: now},
			},
			expectedCount: 3, // Only v3.0.0, v3.0.1, v3.0.2
			expectedRCs:   map[string]int{"3.0.0": 0},
			expectedPatch: map[string]int{"3.0.0": 2},
		},
		{
			name: "no_production_release",
			tags: []tagWithCommit{
				{Name: "v4.0.0-rc1", SHA: "sha1", Timestamp: now.Add(-5 * time.Hour)},
				{Name: "v4.0.0-rc2", SHA: "sha2", Timestamp: now.Add(-4 * time.Hour)},
			},
			expectedCount: 2,
			expectedRCs:   map[string]int{}, // No production release yet
			expectedPatch: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployments := collector.parseTagsToDeployments(tt.tags, now.Add(-24*time.Hour), now.Add(time.Hour))

			if len(deployments) != tt.expectedCount {
				t.Errorf("Expected %d deployments, got %d", tt.expectedCount, len(deployments))
			}

			for version, expectedRCs := range tt.expectedRCs {
				found := false
				for _, d := range deployments {
					if d.Version == version && !d.IsRC && !d.IsPatch {
						found = true
						if d.TotalRCs != expectedRCs {
							t.Errorf("Version %s: expected %d RCs, got %d", version, expectedRCs, d.TotalRCs)
						}
						break
					}
				}
				if !found {
					t.Errorf("Production release for version %s not found", version)
				}
			}

			for version, expectedPatches := range tt.expectedPatch {
				found := false
				for _, d := range deployments {
					if d.Version == version && !d.IsRC && !d.IsPatch {
						found = true
						if d.TotalPatches != expectedPatches {
							t.Errorf("Version %s: expected %d patches, got %d", version, expectedPatches, d.TotalPatches)
						}
						break
					}
				}
				if !found && expectedPatches > 0 {
					t.Errorf("Production release for version %s not found", version)
				}
			}
		})
	}
}
