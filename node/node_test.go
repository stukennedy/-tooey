package node

import "testing"

func TestTextBuilder(t *testing.T) {
	n := Text("hello")
	if n.Type != TextNode {
		t.Fatalf("expected TextNode, got %d", n.Type)
	}
	if n.Props.Text != "hello" {
		t.Fatalf("expected 'hello', got %q", n.Props.Text)
	}
}

func TestRowBuilder(t *testing.T) {
	n := Row(Text("a"), Text("b"))
	if n.Type != RowNode {
		t.Fatalf("expected RowNode")
	}
	if len(n.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(n.Children))
	}
}

func TestChaining(t *testing.T) {
	n := Text("x").WithKey("k1").WithFlex(2).WithFocusable()
	if n.Props.Key != "k1" {
		t.Fatalf("expected key k1")
	}
	if n.Props.FlexWeight != 2 {
		t.Fatalf("expected flex 2")
	}
	if !n.Props.Focusable {
		t.Fatal("expected focusable")
	}
}

func TestSpacer(t *testing.T) {
	n := Spacer()
	if n.Type != SpacerNode {
		t.Fatal("expected SpacerNode")
	}
	if n.Props.FlexWeight != 1 {
		t.Fatal("expected flex weight 1")
	}
}

func TestBox(t *testing.T) {
	n := Box(BorderSingle, Text("content"))
	if n.Type != BoxNode {
		t.Fatal("expected BoxNode")
	}
	if len(n.Children) != 1 {
		t.Fatal("expected 1 child")
	}
	if n.Props.Border != BorderSingle {
		t.Fatal("expected single border")
	}
}
