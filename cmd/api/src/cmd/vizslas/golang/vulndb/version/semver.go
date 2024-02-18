// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package version provides shared utilities for manipulating
// Go semantic versions with no prefix.
package version

import (
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

// IsValid reports whether v is a valid unprefixed semantic version.
func IsValid(v string) bool {
	return semver.IsValid("v" + v)
}

// Before reports whether v < v2, where v and v2 are unprefixed semantic
// versions.
func Before(v, v2 string) bool {
	return semver.Compare("v"+v, "v"+v2) < 0
}

// Major returns the major version (e.g. "v2") of the
// unprefixed semantic version v.
func Major(v string) string {
	return semver.Major("v" + v)
}

// Canonical returns the canonical, unprefixed form of the version v,
// which should be an unprefixed semantic version.
// Unlike semver.Canonical, this function preserves build tags.
func Canonical(v string) string {
	sv := "v" + v
	build := semver.Build(sv)
	c := strings.TrimPrefix(semver.Canonical(sv), "v")
	return c + build
}

// TrimPrefix removes the 'v' or 'go' prefix from the given
// semantic version v.
func TrimPrefix(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "go")
	return v
}

var commitHashRegex = regexp.MustCompile(`^[a-f0-9]+$`)

func IsCommitHash(v string) bool {
	return commitHashRegex.MatchString(v)
}
