package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
}

func ParseVersion(version string) (Version, error) {
	split := strings.Split(version, ".")
	if len(split) != 2 {
		return Version{0, 0}, fmt.Errorf("Invalid version string: '%s'.", version)
	}
	return ParseVersionMajMin(split[0], split[1])
}

func ParseVersionMajMin(major, minor string) (Version, error) {
	majorNumber, err := strconv.Atoi(major)
	if err != nil {
		return Version{0, 0}, fmt.Errorf("Invalid major version number: '%s'.", major)
	}
	minorNumber, err := strconv.Atoi(minor)
	if err != nil {
		return Version{0, 0}, fmt.Errorf("Invalid minor version number: '%s'.", minor)
	}
	return Version{majorNumber, minorNumber}, nil
}

func (v Version) Compare(v2 Version) int {
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

func (v Version) Valid() bool {
	return v.Major != 0 || v.Minor != 0
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}
