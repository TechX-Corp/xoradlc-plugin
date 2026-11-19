package api

import (
	"fmt"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
)

// ParseSemver parses vX.Y.Z or X.Y.Z.
func ParseSemver(tag string) ([3]int, bool) {
	tag = strings.TrimPrefix(tag, "v")
	parts := strings.Split(tag, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var v [3]int
	for i, p := range parts {
		if _, err := fmt.Sscanf(p, "%d", &v[i]); err != nil {
			return [3]int{}, false
		}
	}
	return v, true
}

// SemverGreater returns true when a > b.
func SemverGreater(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

// LatestReleaseNewerThan returns the newest vX.Y.Z tag strictly greater than current, or "".
func LatestReleaseNewerThan(current string) string {
	cur, ok := ParseSemver(current)
	if !ok {
		cur = [3]int{}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=100", configRepo())
	status, out := Request("GET", url, RequestOpts{})
	if status != 200 {
		return ""
	}
	list, ok := out.([]any)
	if !ok {
		return ""
	}
	var bestName string
	var bestVer [3]int
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := fmt.Sprint(m["name"])
		v, ok := ParseSemver(name)
		if !ok {
			continue
		}
		if !SemverGreater(v, cur) {
			continue
		}
		if bestName == "" || SemverGreater(v, bestVer) {
			bestName = name
			bestVer = v
		}
	}
	return bestName
}

func configRepo() string { return config.Repo }