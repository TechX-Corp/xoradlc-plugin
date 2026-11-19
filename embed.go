// Package packembed embeds the plugin pack shipped with the xdlc binary.
package packembed

import "embed"

// Pack is roles/, host-agents asset files (manifest + hooks), and standards/.
// Go sources under host-agents/*/ are compiled packages — not embedded.
//
//go:embed all:roles all:standards
//go:embed host-agents/claude/manifest.yaml
//go:embed all:host-agents/claude/hooks
//go:embed host-agents/kiro/manifest.yaml
//go:embed all:host-agents/kiro/hooks
var Pack embed.FS
