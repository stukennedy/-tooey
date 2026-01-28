package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stukennedy/tooey/cell"
	"github.com/stukennedy/tooey/diff"
	"github.com/stukennedy/tooey/node"
)

func TestRenderSingleChange(t *testing.T) {
	changes := []diff.Change{
		{X: 2, Y: 3, Cells: []cell.Cell{{Rune: 'A'}}},
	}
	var buf bytes.Buffer
	Render(&buf, changes)
	out := buf.String()
	// Should contain cursor move to row 4, col 3 (1-based)
	if !strings.Contains(out, "\x1b[4;3H") {
		t.Fatalf("missing cursor move, got: %q", out)
	}
	if !strings.Contains(out, "A") {
		t.Fatal("missing character A")
	}
}

func TestRenderStyleChange(t *testing.T) {
	changes := []diff.Change{
		{X: 0, Y: 0, Cells: []cell.Cell{
			{Rune: 'A', FG: 1, Style: node.Bold},
			{Rune: 'B', FG: 1, Style: node.Bold}, // same style, no extra SGR
			{Rune: 'C', FG: 2},                    // different, new SGR
		}},
	}
	var buf bytes.Buffer
	Render(&buf, changes)
	out := buf.String()
	// Should have SGR for bold+fg1, then SGR for fg2
	if !strings.Contains(out, ";1;38;5;1m") {
		t.Fatalf("missing bold+fg1 SGR, got: %q", out)
	}
	if !strings.Contains(out, ";38;5;2m") {
		t.Fatalf("missing fg2 SGR, got: %q", out)
	}
	// "B" should NOT have its own SGR (same as A)
	// Count SGR sequences
	count := strings.Count(out, "\x1b[0")
	// 2 style changes + 1 final reset = 3
	if count != 3 {
		t.Fatalf("expected 3 SGR sequences, got %d in %q", count, out)
	}
}

func TestRenderNoChanges(t *testing.T) {
	var buf bytes.Buffer
	Render(&buf, nil)
	if buf.Len() != 0 {
		t.Fatalf("expected empty output, got %q", buf.String())
	}
}
