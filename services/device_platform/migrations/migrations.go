// Package migrations embeds device_platform schema migrations.
package migrations

import "embed"

// Files contains ordered SQL migrations applied during service startup.
//
//go:embed *.sql
var Files embed.FS
