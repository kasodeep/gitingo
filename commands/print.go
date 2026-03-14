// Package commands implements all gitingo commands.
// The package-level printer p is shared across all command files.
package commands

import "github.com/kasodeep/gitingo/printer"

var p = printer.NewPrettyPrinter()
