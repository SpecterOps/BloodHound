// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// SPDX-License-Identifier: Apache-2.0
package prsize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountNumstat(t *testing.T) {
	t.Parallel()

	resultValue, err := countNumstat([]byte("100\t20\tserver/feature.go\x00200\t10\tserver/feature_test.go\x00300\t10\tdocs/feature.md\x00-\t-\t.yarn/cache/example.zip\x00"))

	require.NoError(t, err)
	require.Equal(t, 120, resultValue.changedLines)
	require.Len(t, resultValue.excludedFiles, 2)
	require.Len(t, resultValue.binaryFiles, 1)
	require.Equal(t, "docs/feature.md", resultValue.excludedFiles[0].path)
}

func TestIsExcluded(t *testing.T) {
	t.Parallel()

	require.True(t, isExcluded("server/feature_test.go"))
	require.True(t, isExcluded("docs/feature.md"))
	require.True(t, isExcluded(".yarn/cache/package.zip"))
	require.False(t, isExcluded("server/feature.go"))
}
