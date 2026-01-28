package component

import (
	"strings"

	"github.com/stukennedy/tooey/input"
	"github.com/stukennedy/tooey/node"
)

// TextInput holds state for a multi-line text input with cursor.
type TextInput struct {
	Value       string
	Cursor      int // rune offset into Value
	Placeholder string
	Focused     bool
}

// NewTextInput creates a text input with a placeholder.
func NewTextInput(placeholder string) TextInput {
	return TextInput{Placeholder: placeholder, Focused: true}
}

// Update handles a key event and returns the updated TextInput.
func (ti TextInput) Update(key input.Key) TextInput {
	runes := []rune(ti.Value)
	switch key.Type {
	case input.RuneKey:
		runes = append(runes[:ti.Cursor], append([]rune{key.Rune}, runes[ti.Cursor:]...)...)
		ti.Cursor++
	case input.ShiftEnter:
		runes = append(runes[:ti.Cursor], append([]rune{'\n'}, runes[ti.Cursor:]...)...)
		ti.Cursor++
	case input.Backspace:
		if ti.Cursor > 0 {
			runes = append(runes[:ti.Cursor-1], runes[ti.Cursor:]...)
			ti.Cursor--
		}
	case input.Delete:
		if ti.Cursor < len(runes) {
			runes = append(runes[:ti.Cursor], runes[ti.Cursor+1:]...)
		}
	case input.Left:
		if ti.Cursor > 0 {
			ti.Cursor--
		}
	case input.Right:
		if ti.Cursor < len(runes) {
			ti.Cursor++
		}
	case input.Home:
		// Move to start of current line
		ti.Cursor = lineStart(runes, ti.Cursor)
	case input.End:
		// Move to end of current line
		ti.Cursor = lineEnd(runes, ti.Cursor)
	case input.Up:
		ti.Cursor = moveCursorUp(runes, ti.Cursor)
	case input.Down:
		ti.Cursor = moveCursorDown(runes, ti.Cursor)
	}
	ti.Value = string(runes)
	return ti
}

// Submit returns the current value and resets the input.
func (ti TextInput) Submit() (string, TextInput) {
	val := strings.TrimSpace(ti.Value)
	ti.Value = ""
	ti.Cursor = 0
	return val, ti
}

// LineCount returns the number of display lines.
func (ti TextInput) LineCount() int {
	if ti.Value == "" {
		return 1
	}
	return strings.Count(ti.Value, "\n") + 1
}

// Render returns a node tree displaying the multi-line input with cursor.
func (ti TextInput) Render(prefix string, fg, bg node.Color) node.Node {
	if ti.Value == "" {
		// Show cursor block + placeholder when focused and empty
		if ti.Focused {
			return node.Row(
				node.TextStyled(prefix, fg, bg, 0),
				node.TextStyled(" ", node.Color(0), node.Color(15), 0), // block cursor
				node.TextStyled(ti.Placeholder, node.Color(8), bg, node.Dim),
			)
		}
		return node.TextStyled(prefix+ti.Placeholder, node.Color(8), bg, node.Dim)
	}

	runes := []rune(ti.Value)
	lines := splitLines(string(runes))

	// Find which line the cursor is on and its column offset
	cursorLine, cursorCol := cursorPosition(runes, ti.Cursor)

	// Continuation lines indent to align with text after prefix
	contPrefix := strings.Repeat(" ", len([]rune(prefix)))

	var lineNodes []node.Node
	for i, line := range lines {
		var ln node.Node
		linePrefix := contPrefix
		if i == 0 {
			linePrefix = prefix
		}

		if i == cursorLine && ti.Focused {
			// This line has the cursor
			lineRunes := []rune(line)
			before := string(lineRunes[:cursorCol])
			var cursorChar string
			var after string
			if cursorCol < len(lineRunes) {
				cursorChar = string(lineRunes[cursorCol])
				after = string(lineRunes[cursorCol+1:])
			} else {
				cursorChar = " "
			}
			ln = node.Row(
				node.TextStyled(linePrefix+before, fg, bg, 0),
				node.TextStyled(cursorChar, node.Color(0), node.Color(15), 0),
				node.TextStyled(after, fg, bg, 0),
			)
		} else {
			ln = node.TextStyled(linePrefix+line, fg, bg, 0)
		}
		lineNodes = append(lineNodes, ln)
	}

	if len(lineNodes) == 1 {
		return lineNodes[0]
	}
	return node.Column(lineNodes...)
}

// splitLines splits on newline, always returning at least one element.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := strings.Split(s, "\n")
	return lines
}

// cursorPosition converts a flat rune offset to (line, col).
func cursorPosition(runes []rune, cursor int) (int, int) {
	line, col := 0, 0
	for i := 0; i < cursor && i < len(runes); i++ {
		if runes[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

// lineStart returns the rune index of the start of the current line.
func lineStart(runes []rune, cursor int) int {
	for i := cursor - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// lineEnd returns the rune index of the end of the current line.
func lineEnd(runes []rune, cursor int) int {
	for i := cursor; i < len(runes); i++ {
		if runes[i] == '\n' {
			return i
		}
	}
	return len(runes)
}

// moveCursorUp moves the cursor to the same column on the previous line.
func moveCursorUp(runes []rune, cursor int) int {
	_, col := cursorPosition(runes, cursor)
	start := lineStart(runes, cursor)
	if start == 0 {
		return 0 // already on first line
	}
	// Go to previous line
	prevLineEnd := start - 1 // the \n char
	prevLineStart := lineStart(runes, prevLineEnd)
	prevLineLen := prevLineEnd - prevLineStart
	if col > prevLineLen {
		col = prevLineLen
	}
	return prevLineStart + col
}

// moveCursorDown moves the cursor to the same column on the next line.
func moveCursorDown(runes []rune, cursor int) int {
	_, col := cursorPosition(runes, cursor)
	end := lineEnd(runes, cursor)
	if end >= len(runes) {
		return len(runes) // already on last line
	}
	// Go to next line
	nextLineStart := end + 1 // skip the \n
	nextLineEnd := lineEnd(runes, nextLineStart)
	nextLineLen := nextLineEnd - nextLineStart
	if col > nextLineLen {
		col = nextLineLen
	}
	return nextLineStart + col
}
