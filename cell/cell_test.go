package cell

import "testing"

func TestNewBuffer(t *testing.T) {
	b := NewBuffer(10, 5)
	if b.Width != 10 || b.Height != 5 {
		t.Fatalf("wrong size")
	}
	if len(b.Cells) != 50 {
		t.Fatalf("expected 50 cells, got %d", len(b.Cells))
	}
	// All cells should be spaces
	for i, c := range b.Cells {
		if c.Rune != ' ' {
			t.Fatalf("cell %d not space", i)
		}
	}
}

func TestSetGet(t *testing.T) {
	b := NewBuffer(5, 5)
	c := Cell{Rune: 'X', FG: 1, BG: 2}
	b.Set(2, 3, c)
	got := b.Get(2, 3)
	if got != c {
		t.Fatalf("expected %v, got %v", c, got)
	}
}

func TestOutOfBounds(t *testing.T) {
	b := NewBuffer(3, 3)
	b.Set(-1, 0, Cell{Rune: 'X'}) // should not panic
	b.Set(3, 0, Cell{Rune: 'X'})  // should not panic
	got := b.Get(-1, 0)
	if got.Rune != 0 {
		t.Fatal("expected zero cell for out of bounds")
	}
}

func TestWriteString(t *testing.T) {
	b := NewBuffer(10, 1)
	b.WriteString(2, 0, "hi", 1, 0, 0)
	if b.Get(2, 0).Rune != 'h' {
		t.Fatal("expected 'h' at (2,0)")
	}
	if b.Get(3, 0).Rune != 'i' {
		t.Fatal("expected 'i' at (3,0)")
	}
	if b.Get(4, 0).Rune != ' ' {
		t.Fatal("expected space at (4,0)")
	}
}

func TestWriteStringClips(t *testing.T) {
	b := NewBuffer(3, 1)
	b.WriteString(1, 0, "abcdef", 0, 0, 0) // only 2 chars fit
	if b.Get(1, 0).Rune != 'a' {
		t.Fatal("expected 'a'")
	}
	if b.Get(2, 0).Rune != 'b' {
		t.Fatal("expected 'b'")
	}
}

func TestClear(t *testing.T) {
	b := NewBuffer(3, 3)
	b.Set(1, 1, Cell{Rune: 'Z', FG: 5})
	b.Clear()
	got := b.Get(1, 1)
	if got.Rune != ' ' || got.FG != 0 {
		t.Fatal("clear didn't reset cell")
	}
}
