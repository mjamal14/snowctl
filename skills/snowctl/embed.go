// Package snowctlskill embeds the full snowctl skill content (SKILL.md and all
// reference files) so that other packages can access it without filesystem
// reads or build-time code generation.
package snowctlskill

import "embed"

// Content is the embedded filesystem rooted at skills/snowctl/.
// It contains SKILL.md and the references/ subtree.
//
//go:embed SKILL.md references
var Content embed.FS
