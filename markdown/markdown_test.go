package markdown

import (
	"strings"
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestEmptyInput(t *testing.T) {
	nodes := Render("", 40, 0)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node for empty input, got %d", len(nodes))
	}
}

func TestHeading(t *testing.T) {
	nodes := Render("# Hello", 40, 7)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Type != node.RowNode {
		t.Fatalf("expected RowNode, got %v", n.Type)
	}
	if len(n.Children) < 1 {
		t.Fatal("expected children")
	}
	if n.Children[0].Props.Style&node.Bold == 0 {
		t.Error("heading should be bold")
	}
	if n.Children[0].Props.Text != "Hello" {
		t.Errorf("expected 'Hello', got %q", n.Children[0].Props.Text)
	}
}

func TestHeadingLevels(t *testing.T) {
	for _, tc := range []string{"## Sub", "### Sub"} {
		nodes := Render(tc, 40, 7)
		if len(nodes) != 1 {
			t.Errorf("expected 1 node for %q", tc)
		}
	}
}

func TestBold(t *testing.T) {
	nodes := Render("**bold**", 40, 7)
	if len(nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	child := nodes[0].Children[0]
	if child.Props.Style&node.Bold == 0 {
		t.Error("expected bold")
	}
	if child.Props.Text != "bold" {
		t.Errorf("expected 'bold', got %q", child.Props.Text)
	}
}

func TestItalic(t *testing.T) {
	nodes := Render("*italic*", 40, 7)
	child := nodes[0].Children[0]
	if child.Props.Style&node.Italic == 0 {
		t.Error("expected italic")
	}
}

func TestBoldItalic(t *testing.T) {
	nodes := Render("***both***", 40, 7)
	child := nodes[0].Children[0]
	if child.Props.Style&(node.Bold|node.Italic) != node.Bold|node.Italic {
		t.Error("expected bold+italic")
	}
}

func TestInlineCode(t *testing.T) {
	nodes := Render("use `foo` here", 40, 7)
	row := nodes[0]
	if len(row.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(row.Children))
	}
	if row.Children[1].Props.Text != "foo" {
		t.Errorf("expected 'foo', got %q", row.Children[1].Props.Text)
	}
	if row.Children[1].Props.BG == 0 {
		t.Error("expected code background color")
	}
}

func TestLink(t *testing.T) {
	nodes := Render("[click](http://x)", 40, 7)
	child := nodes[0].Children[0]
	if child.Props.Text != "click" {
		t.Errorf("expected 'click', got %q", child.Props.Text)
	}
	if child.Props.Style&node.Underline == 0 {
		t.Error("expected underline for link")
	}
}

func TestBulletList(t *testing.T) {
	nodes := Render("- item one\n- item two", 40, 7)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if !strings.Contains(nodes[0].Children[0].Props.Text, "•") {
		t.Error("expected bullet character")
	}
}

func TestNumberedList(t *testing.T) {
	nodes := Render("1. first\n2. second", 40, 7)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if !strings.Contains(nodes[0].Children[0].Props.Text, "1.") {
		t.Error("expected number prefix")
	}
}

func TestBlockquote(t *testing.T) {
	nodes := Render("> quoted text", 40, 7)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	row := nodes[0]
	if !strings.Contains(row.Children[0].Props.Text, "│") {
		t.Error("expected quote bar")
	}
	// Content should be italic
	if row.Children[1].Props.Style&node.Italic == 0 {
		t.Error("expected italic in blockquote")
	}
}

func TestCodeBlock(t *testing.T) {
	input := "```go\nfmt.Println()\n```"
	nodes := Render(input, 40, 7)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != node.BoxNode {
		t.Errorf("expected BoxNode, got %v", nodes[0].Type)
	}
}

func TestCodeBlockNoLang(t *testing.T) {
	input := "```\ncode here\n```"
	nodes := Render(input, 40, 7)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != node.BoxNode {
		t.Errorf("expected BoxNode, got %v", nodes[0].Type)
	}
}

func TestHorizontalRule(t *testing.T) {
	nodes := Render("---", 20, 7)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if len(nodes[0].Props.Text) != 20*len("─") {
		// ─ is 3 bytes in UTF-8
		expected := strings.Repeat("─", 20)
		if nodes[0].Props.Text != expected {
			t.Errorf("expected %d-wide rule, got %q", 20, nodes[0].Props.Text)
		}
	}
}

func TestCheckboxChecked(t *testing.T) {
	nodes := Render("- [x] done", 40, 7)
	if len(nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	if !strings.Contains(nodes[0].Children[0].Props.Text, "✔") {
		t.Error("expected check mark")
	}
}

func TestCheckboxUnchecked(t *testing.T) {
	nodes := Render("- [ ] todo", 40, 7)
	if len(nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	if !strings.Contains(nodes[0].Children[0].Props.Text, "☐") {
		t.Error("expected unchecked box")
	}
	// Text should be dim
	if nodes[0].Children[1].Props.Style&node.Dim == 0 {
		t.Error("expected dim style for unchecked item")
	}
}

func TestUnclosedMarkers(t *testing.T) {
	// Unclosed bold marker should be treated as literal
	nodes := Render("**unclosed", 40, 7)
	if len(nodes) != 1 {
		t.Fatal("expected 1 node")
	}
	child := nodes[0].Children[0]
	if child.Props.Text != "**unclosed" {
		t.Errorf("expected literal '**unclosed', got %q", child.Props.Text)
	}
}

func TestCombinedStyles(t *testing.T) {
	nodes := Render("hello **bold** and *italic*", 40, 7)
	row := nodes[0]
	if len(row.Children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(row.Children))
	}
	if row.Children[1].Props.Style&node.Bold == 0 {
		t.Error("expected bold")
	}
	if row.Children[3].Props.Style&node.Italic == 0 {
		t.Error("expected italic")
	}
}
