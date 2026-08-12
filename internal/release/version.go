package release

import (
	"fmt"
	"strconv"
	"strings"
)

type semver struct {
	Major, Minor, Patch int
}

func parseSemver(s string) (semver, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("release: invalid version %q (expected X.Y.Z)", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, fmt.Errorf("release: invalid major in %q: %w", s, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, fmt.Errorf("release: invalid minor in %q: %w", s, err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, fmt.Errorf("release: invalid patch in %q: %w", s, err)
	}
	return semver{major, minor, patch}, nil
}

func (v semver) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// BumpVersion takes a current version string and a bump type ("major", "minor", "patch")
// and returns the bumped version as "vX.Y.Z".
func BumpVersion(current, bump string) (string, error) {
	v, err := parseSemver(current)
	if err != nil {
		return "", err
	}
	switch bump {
	case "major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case "minor":
		v.Minor++
		v.Patch = 0
	case "patch":
		v.Patch++
	default:
		return "", fmt.Errorf("release: unknown bump type %q (use major, minor, or patch)", bump)
	}
	return v.String(), nil
}
