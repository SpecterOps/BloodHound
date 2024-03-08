package git_test

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/stbernard/git"
	"github.com/stretchr/testify/require"
)

func TestParseLatestVersionFromTags(t *testing.T) {
	_, err := git.ParseLatestVersionFromTags(".", []string{})
	require.Nil(t, err)
}
