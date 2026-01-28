package cell

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func TestPaintTextInRow(t *testing.T) {
	tree := node.Row(node.Text("ab"), node.Text("cd"))
	lt := layout.Layout(tree, 10, 1)
	buf := NewBuffer(10, 1)
	Paint(buf, lt)

	expected := "ab"
	for i, ch := range expected {
		if buf.Get(i, 0).Rune != ch {
			t.Fatalf("pos %d: expected %c, got %c", i, ch, buf.Get(i, 0).Rune)
		}
	}
	// "cd" starts at x=2
	if buf.Get(2, 0).Rune != 'c' || buf.Get(3, 0).Rune != 'd' {
		t.Fatalf("expected 'cd' at x=2,3, got %c%c", buf.Get(2, 0).Rune, buf.Get(3, 0).Rune)
	}
}

func TestPaintBox(t *testing.T) {
	tree := node.Box(node.BorderSingle, node.Text("hi"))
	lt := layout.Layout(tree, 10, 5)
	buf := NewBuffer(10, 5)
	Paint(buf, lt)

	if buf.Get(0, 0).Rune != '┌' {
		t.Fatalf("top-left: expected ┌, got %c", buf.Get(0, 0).Rune)
	}
	if buf.Get(9, 0).Rune != '┐' {
		t.Fatalf("top-right: expected ┐, got %c", buf.Get(9, 0).Rune)
	}
	if buf.Get(0, 4).Rune != '└' {
		t.Fatalf("bottom-left: expected └, got %c", buf.Get(0, 4).Rune)
	}
	// Inner text "hi" at (1,1)
	if buf.Get(1, 1).Rune != 'h' || buf.Get(2, 1).Rune != 'i' {
		t.Fatalf("inner text: expected 'hi', got %c%c", buf.Get(1, 1).Rune, buf.Get(2, 1).Rune)
	}
}

func TestPaintColumn(t *testing.T) {
	tree := node.Column(node.Text("top"), node.Text("bot"))
	lt := layout.Layout(tree, 10, 5)
	buf := NewBuffer(10, 5)
	Paint(buf, lt)

	if buf.Get(0, 0).Rune != 't' {
		t.Fatalf("expected 't' at (0,0), got %c", buf.Get(0, 0).Rune)
	}
	if buf.Get(0, 1).Rune != 'b' {
		t.Fatalf("expected 'b' at (0,1), got %c", buf.Get(0, 1).Rune)
	}
}
