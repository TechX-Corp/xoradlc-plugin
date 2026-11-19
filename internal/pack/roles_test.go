package pack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogRolesFromDir(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "roles", "catalog.yaml")); err != nil {
		t.Skip("catalog.yaml not at module root")
	}
	os.Setenv("XDLC_PLUGIN_DIR", root)
	t.Cleanup(func() { os.Unsetenv("XDLC_PLUGIN_DIR") })
	roles, err := CatalogRoles("")
	if err != nil {
		t.Fatal(err)
	}
	var hasDev, hasBA bool
	for _, r := range roles {
		if r.ID == "dev" && r.Enabled {
			hasDev = true
		}
		if r.ID == "ba" && !r.Enabled {
			hasBA = true
		}
	}
	if !hasDev || !hasBA {
		t.Fatalf("expected enabled dev and disabled ba, got %#v", roles)
	}
}
