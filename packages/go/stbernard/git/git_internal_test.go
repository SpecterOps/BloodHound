package git

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseLatestVersion(t *testing.T) {
	latestRelease := *semver.MustParse("v5.5.0")
	latestRC := *semver.MustParse("v5.5.0-rc1")

	t.Run("empty list of versions", func(t *testing.T) {
		_, err := parseLatestVersion([]string{})
		require.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("nil list of versions", func(t *testing.T) {
		_, err := parseLatestVersion(nil)
		assert.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("list of one invalid version", func(t *testing.T) {
		_, err := parseLatestVersion([]string{"invalid"})
		assert.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("list of valid versions with one invalid version", func(t *testing.T) {
		latest, err := parseLatestVersion([]string{"v5.3.0-rc1", "v5.5.0-rc1", "v5.4.0", "invalid"})
		require.Nil(t, err)
		assert.Equal(t, latestRC, latest)
	})

	t.Run("list of valid versions with latest non-RC", func(t *testing.T) {
		latest, err := parseLatestVersion([]string{"v5.3.0-rc1", "v5.5.0", "v5.4.0", "v5.4.0-rc1"})
		require.Nil(t, err)
		assert.Equal(t, latestRelease, latest)
	})
}
