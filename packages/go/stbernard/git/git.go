package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/log"
)

var (
	ErrNoValidSemverFound = errors.New("no valid semver found")
)

// ParseLatestVersionFromTags gets the latest semver tag in the repository
func ParseLatestVersionFromTags(path string, env []string) (semver.Version, error) {
	var (
		version semver.Version
	)

	if versions, err := getAllVersionTags(path, env); err != nil {
		return version, fmt.Errorf("get version tags from git: %w", err)
	} else {
		return parseLatestVersion(versions)
	}
}

// parseLatestVersion parses a list of found versions and returns the latest from among them
func parseLatestVersion(versions []string) (semver.Version, error) {
	if len(versions) == 0 {
		return semver.Version{}, ErrNoValidSemverFound
	}

	semversions := make([]*semver.Version, 0, len(versions))

	for _, version := range versions {
		if v, err := semver.NewVersion(version); err != nil {
			// skip if version string isn't valid
			continue
		} else {
			semversions = append(semversions, v)
		}
	}

	if len(semversions) == 0 {
		return semver.Version{}, ErrNoValidSemverFound
	}

	sort.Sort(semver.Collection(semversions))

	return *semversions[len(semversions)-1], nil
}

// getAllVersionTags gets the version tags from git and dumps them into a []string
func getAllVersionTags(path string, env []string) ([]string, error) {
	var (
		output bytes.Buffer
	)

	cmd := exec.Command("git", "tag", "--list", "v*")
	cmd.Env = env
	cmd.Dir = path
	cmd.Stdout = &output

	if log.GlobalAccepts(log.LevelDebug) {
		cmd.Stderr = os.Stderr
	}

	log.Infof("Listing tags for %v", path)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git tag --list v*: %w", err)
	}

	log.Infof("Finished listing tags for %v", path)

	return strings.Split(output.String(), "\n"), nil
}
