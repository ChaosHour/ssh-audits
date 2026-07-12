package sshutil

import (
	"bytes"
	"sync"
	"testing"

	"github.com/fatih/color"
)

func TestLinePrinterPadsHostNames(t *testing.T) {
	old := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = old })

	var buf bytes.Buffer
	p := newLinePrinter(&buf, []string{"db1", "primary"})
	p.print("db1", "hello")
	p.print("primary", "world")

	want := "[db1]     hello\n[primary] world\n"
	if got := buf.String(); got != want {
		t.Errorf("printer output:\n%q\nwant:\n%q", got, want)
	}
}

func TestLinePrinterKeepsLinesWhole(t *testing.T) {
	old := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = old })

	var buf bytes.Buffer
	p := newLinePrinter(&buf, []string{"a", "b"})

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { p.print("a", "aaaa") })
		wg.Go(func() { p.print("b", "bbbb") })
	}
	wg.Wait()

	for i, line := range bytes.Split(bytes.TrimRight(buf.Bytes(), "\n"), []byte("\n")) {
		s := string(line)
		if s != "[a] aaaa" && s != "[b] bbbb" {
			t.Fatalf("line %d interleaved or malformed: %q", i, s)
		}
	}
}
