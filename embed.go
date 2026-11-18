// Package packembed embeds the plugin pack shipped with the xdlc binary.
package packembed

import "embed"

// Pack is roles/, host-agents/, and standards/ at the module root.
//
//go:embed all:roles all:host-agents all:standards
var Pack embed.FS
