package main

import (
	"fmt"
	"strconv"
	"strings"
)

// A Version wraps a major and minor integer version pair.
type Version struct {
	Major int
	Minor int
}

// ParseVersion returns a Version from a "major.minor" or "all" version string.
func ParseVersion(version string) (Version, error) {
	if version == "all" {
		return Version{-1, -1}, nil
	}

	split := strings.Split(version, ".")
	if len(split) != 2 {
		return Version{0, 0}, fmt.Errorf("invalid version string: %s", version)
	}
	majorNumber, err := strconv.Atoi(split[0])
	if err != nil {
		return Version{0, 0}, fmt.Errorf("invalid major version number: %s", split[0])
	}
	minorNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return Version{0, 0}, fmt.Errorf("invalid minor version number: %s", split[1])
	}
	return Version{majorNumber, minorNumber}, nil
}

// Compare compares two versions, returning 1, 0, or -1 if the compared version
// is before, after, or equal to this version respectively. The "all versions"
// pseudo-version is equal to all other versions.
func (v Version) Compare(v2 Version) int {
	if v.IsAll() || v2.IsAll() {
		return 0
	}
	if v.Major < v2.Major {
		return -1
	} else if v.Major > v2.Major {
		return 1
	} else if v.Minor < v2.Minor {
		return -1
	} else if v.Minor > v2.Minor {
		return 1
	}
	return 0
}

// IsAll returns true if this version is the "all versions" pseudo-version.
func (v Version) IsAll() bool {
	return v.Major < 0
}

func (v Version) String() string {
	if v.IsAll() {
		return "all"
	}
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}
