// utility/version.go
package Utility

import (
	"strings"
)

// Base on https://go.dev/doc/modules/version-numbers for version number
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

func NewVersion(str string) *Version {
	v := new(Version)
	v.Parse(str)
	return v
}

// Parse values from string (e.g., "v1.2.3" or "v1.2.3-beta.1")
func (v *Version) Parse(str string) {
	values := strings.Split(str, ".")
	if len(values) < 3 {
		// fallback to zeros on malformed strings
		v.Major, v.Minor, v.Patch, v.PreRelease = 0, 0, 0, ""
		return
	}

	v.Major = ToInt(strings.ReplaceAll(values[0], "v", ""))
	v.Minor = ToInt(values[1])

	// handle patch + optional pre-release
	if strings.Contains(values[2], "-") {
		parts := strings.SplitN(values[2], "-", 2)
		v.Patch = ToInt(parts[0])
		if len(parts) == 2 {
			v.PreRelease = parts[1]
		}
	} else {
		v.Patch = ToInt(values[2])
	}
}

// Stringnify the version.
func (v *Version) ToString() string {
	str := "v" + ToString(v.Major) + "." + ToString(v.Minor) + "." + ToString(v.Patch)
	if len(v.PreRelease) > 0 {
		str += "-" + v.PreRelease
	}
	return str
}

// Compare two versions: 1 means v is newer than 'to', 0 is the same, -1 is older.
// PreRelease is NOT compared (treated as informational only).
func (v *Version) Compare(to *Version) int {
	if v.Major > to.Major {
		return 1
	} else if v.Major < to.Major {
		return -1
	}

	if v.Minor > to.Minor {
		return 1
	} else if v.Minor < to.Minor {
		return -1
	}

	if v.Patch > to.Patch {
		return 1
	} else if v.Patch < to.Patch {
		return -1
	}

	// here all info are equal; the Pre-Release info is not comparable...
	return 0
}

