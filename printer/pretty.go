// Package printer provides a shared output interface used by all commands.
package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/text"
)

// Printer is the output contract every command depends on.
// Keeping it an interface lets commands be tested with a fake writer.
type Printer interface {
	Info(msg string)
	Success(msg string)
	Warn(msg string)
	Error(msg string)
}

// PrettyPrinter is the production implementation of Printer.
// It writes coloured output to stdout and errors to stderr.
type PrettyPrinter struct {
	out io.Writer
	err io.Writer
}

func NewPrettyPrinter() *PrettyPrinter {
	return &PrettyPrinter{out: os.Stdout, err: os.Stderr}
}

func (p *PrettyPrinter) Info(msg string)    { fmt.Fprintln(p.out, msg) }
func (p *PrettyPrinter) Success(msg string) { fmt.Fprintln(p.out, text.FgGreen.Sprint(msg)) }
func (p *PrettyPrinter) Warn(msg string)    { fmt.Fprintln(p.out, text.FgYellow.Sprint(msg)) }
func (p *PrettyPrinter) Error(msg string)   { fmt.Fprintln(p.err, text.FgRed.Sprint(msg)) }

// ─────────────────────────────────────────────────────────────────────────────
// Format helpers — used by log output, not part of the Printer interface
// because they return strings rather than writing directly.
// ─────────────────────────────────────────────────────────────────────────────

func (p *PrettyPrinter) CommitHash(hash string) string { return text.FgYellow.Sprint(hash) }
func (p *PrettyPrinter) Branch(name string) string     { return text.FgCyan.Sprint(name) }
func (p *PrettyPrinter) Author(name, email string) string {
	return text.FgGreen.Sprintf("%s <%s>", name, email)
}
func (p *PrettyPrinter) Date(date string) string   { return text.FgBlue.Sprint(date) }
func (p *PrettyPrinter) Message(msg string) string { return text.FgWhite.Sprint(msg) }
