package diff

import (
	"testing"

	"github.com/stukennedy/tooey/cell"
)

func TestIdenticalBuffers(t *testing.T) {
	a := cell.NewBuffer(10, 5)
	b := cell.NewBuffer(10, 5)
	changes := Diff(a, b)
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes, got %d", len(changes))
	}
}

func TestSingleCellChange(t *testing.T) {
	a := cell.NewBuffer(5, 1)
	b := cell.NewBuffer(5, 1)
	b.Set(2, 0, cell.Cell{Rune: 'X'})
	changes := Diff(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].X != 2 || changes[0].Y != 0 {
		t.Fatalf("wrong position: (%d,%d)", changes[0].X, changes[0].Y)
	}
	if len(changes[0].Cells) != 1 || changes[0].Cells[0].Rune != 'X' {
		t.Fatal("wrong cell content")
	}
}

func TestAdjacentChanges(t *testing.T) {
	a := cell.NewBuffer(5, 1)
	b := cell.NewBuffer(5, 1)
	b.Set(1, 0, cell.Cell{Rune: 'A'})
	b.Set(2, 0, cell.Cell{Rune: 'B'})
	b.Set(3, 0, cell.Cell{Rune: 'C'})
	changes := Diff(a, b)
	if len(changes) != 1 {
		t.Fatalf("expected 1 grouped change, got %d", len(changes))
	}
	if len(changes[0].Cells) != 3 {
		t.Fatalf("expected 3 cells in run, got %d", len(changes[0].Cells))
	}
}

func TestSplitChanges(t *testing.T) {
	a := cell.NewBuffer(10, 1)
	b := cell.NewBuffer(10, 1)
	b.Set(1, 0, cell.Cell{Rune: 'A'})
	b.Set(5, 0, cell.Cell{Rune: 'B'})
	changes := Diff(a, b)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
}
