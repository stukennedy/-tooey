package focus

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func makeTree() layout.LayoutNode {
	tree := node.Column(
		node.Text("a").WithKey("a").WithFocusable(),
		node.Text("b").WithKey("b").WithFocusable(),
		node.Text("c").WithKey("c").WithFocusable(),
	)
	return layout.Layout(tree, 80, 24)
}

func TestTabCycle(t *testing.T) {
	m := NewManager()
	m.Update(makeTree())
	if m.Current() != "a" {
		t.Fatalf("expected 'a', got %q", m.Current())
	}
	m.Next()
	if m.Current() != "b" {
		t.Fatalf("expected 'b', got %q", m.Current())
	}
	m.Next()
	if m.Current() != "c" {
		t.Fatalf("expected 'c', got %q", m.Current())
	}
	m.Next() // wraps
	if m.Current() != "a" {
		t.Fatalf("expected 'a' after wrap, got %q", m.Current())
	}
}

func TestShiftTabCycle(t *testing.T) {
	m := NewManager()
	m.Update(makeTree())
	m.Prev() // wraps to end
	if m.Current() != "c" {
		t.Fatalf("expected 'c', got %q", m.Current())
	}
}

func TestContextPushPop(t *testing.T) {
	m := NewManager()
	m.Update(makeTree())
	m.Next() // on "b"

	// Push into a pane context
	paneTree := node.Column(
		node.Text("x").WithKey("x").WithFocusable(),
		node.Text("y").WithKey("y").WithFocusable(),
	)
	paneLT := layout.Layout(paneTree, 80, 24)
	m.PushContext(paneLT)

	if m.Current() != "x" {
		t.Fatalf("expected 'x' in pane, got %q", m.Current())
	}
	if m.FocusableCount() != 2 {
		t.Fatalf("expected 2 focusables in pane, got %d", m.FocusableCount())
	}

	m.PopContext()
	if m.Current() != "b" {
		t.Fatalf("expected 'b' after pop, got %q", m.Current())
	}
}

func TestEmptyFocusables(t *testing.T) {
	m := NewManager()
	tree := node.Text("no focus")
	lt := layout.Layout(tree, 80, 24)
	m.Update(lt)
	if m.Current() != "" {
		t.Fatalf("expected empty, got %q", m.Current())
	}
	m.Next() // should not panic
	m.Prev() // should not panic
}
