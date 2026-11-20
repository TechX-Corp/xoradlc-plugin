package installs

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const setupIDBytes = 3 // 6 hex chars

// NewID returns a new unique setup id not in existing.
func NewID(existing []Record) string {
	seen := idsInUse(existing)
	return newSetupID(seen)
}

// ByID finds a setup by id (case-insensitive).
func ByID(items []Record, id string) (Record, bool) {
	id = strings.TrimSpace(id)
	for _, r := range items {
		if strings.EqualFold(r.ID, id) {
			return r, true
		}
	}
	return Record{}, false
}

func idsInUse(items []Record) map[string]struct{} {
	seen := map[string]struct{}{}
	for _, r := range items {
		if r.ID != "" {
			seen[strings.ToLower(r.ID)] = struct{}{}
		}
	}
	return seen
}

func ensureIDs(items []Record) ([]Record, bool) {
	seen := idsInUse(items)
	changed := false
	for i := range items {
		if items[i].ID != "" {
			continue
		}
		items[i].ID = newSetupID(seen)
		changed = true
	}
	return items, changed
}

func newSetupID(seen map[string]struct{}) string {
	for {
		var b [setupIDBytes]byte
		_, _ = rand.Read(b[:])
		id := hex.EncodeToString(b[:])
		if _, ok := seen[strings.ToLower(id)]; ok {
			continue
		}
		seen[strings.ToLower(id)] = struct{}{}
		return id
	}
}