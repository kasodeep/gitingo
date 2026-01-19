package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type PrettyPrinter struct {
	out io.Writer
	err io.Writer
}

func NewPrettyPrinter() *PrettyPrinter {
	return &PrettyPrinter{
		out: os.Stdout,
		err: os.Stderr,
	}
}

func (p *PrettyPrinter) Info(msg string) {
	fmt.Fprintln(p.out, msg)
}

func (p *PrettyPrinter) Success(msg string) {
	fmt.Fprintln(p.out, text.FgGreen.Sprint(msg))
}

func (p *PrettyPrinter) Warn(msg string) {
	fmt.Fprintln(p.out, text.FgYellow.Sprint(msg))
}

func (p *PrettyPrinter) Error(msg string) {
	fmt.Fprintln(p.err, text.FgRed.Sprint(msg))
}

func (p *PrettyPrinter) Table(headers []string, rows [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(p.out)

	headerRow := table.Row{}
	for _, h := range headers {
		headerRow = append(headerRow, text.Bold.Sprint(h))
	}
	t.AppendHeader(headerRow)

	for _, r := range rows {
		row := table.Row{}
		for _, c := range r {
			row = append(row, c)
		}
		t.AppendRow(row)
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func (p *PrettyPrinter) CommitHash(hash string) string {
	return text.FgYellow.Sprint(hash)
}

func (p *PrettyPrinter) Branch(name string) string {
	return text.FgCyan.Sprint(name)
}

func (p *PrettyPrinter) Author(name, email string) string {
	return text.FgGreen.Sprintf("%s <%s>", name, email)
}

func (p *PrettyPrinter) Date(date string) string {
	return text.FgBlue.Sprint(date)
}

func (p *PrettyPrinter) Message(msg string) string {
	return text.FgWhite.Sprint(msg)
}
