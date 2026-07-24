package utils

import (
	"strconv"
	"strings"
)

// CompareVersions parses and compares two version strings.
// It returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//    1 if v1 > v2
// It assumes version strings are dot-separated integers (e.g., "1.2.3" or "2.0").
func CompareVersions(v1, v2 string) int {
	v1 = strings.TrimSpace(v1)
	v2 = strings.TrimSpace(v2)

	// Remove leading 'v' if present (e.g., "v1.0.0" -> "1.0.0")
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}
	return 0
}

// CheckVersionConstraint checks if the requestVersion satisfies the constraint: requestVersion [operator] targetVersion
func CheckVersionConstraint(requestVersion, operator, targetVersion string) bool {
	operator = strings.TrimSpace(operator)
	if operator == "" || operator == "*" || operator == "all" {
		return true
	}
	cmp := CompareVersions(requestVersion, targetVersion)
	switch operator {
	case "=":
		return cmp == 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	default:
		return false
	}
}
