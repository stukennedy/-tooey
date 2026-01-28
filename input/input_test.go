package input

import "testing"

func TestParseArrowKeys(t *testing.T) {
	tests := []struct {
		input    []byte
		expected KeyType
	}{
		{[]byte{0x1b, '[', 'A'}, Up},
		{[]byte{0x1b, '[', 'B'}, Down},
		{[]byte{0x1b, '[', 'C'}, Right},
		{[]byte{0x1b, '[', 'D'}, Left},
	}
	for _, tt := range tests {
		keys := parseInput(tt.input)
		if len(keys) != 1 || keys[0].Type != tt.expected {
			t.Errorf("input %v: expected %d, got %v", tt.input, tt.expected, keys)
		}
	}
}

func TestParseRunes(t *testing.T) {
	keys := parseInput([]byte("abc"))
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	for i, ch := range "abc" {
		if keys[i].Type != RuneKey || keys[i].Rune != ch {
			t.Errorf("key %d: expected %c, got %v", i, ch, keys[i])
		}
	}
}

func TestParseSpecialKeys(t *testing.T) {
	tests := []struct {
		input    []byte
		expected KeyType
	}{
		{[]byte{'\r'}, Enter},
		{[]byte{'\t'}, Tab},
		{[]byte{0x1b, '[', 'Z'}, ShiftTab},
		{[]byte{0x7f}, Backspace},
		{[]byte{0x03}, CtrlC},
		{[]byte{0x1b}, Escape},
	}
	for _, tt := range tests {
		keys := parseInput(tt.input)
		if len(keys) != 1 || keys[0].Type != tt.expected {
			t.Errorf("input %v: expected %d, got %v", tt.input, tt.expected, keys)
		}
	}
}

func TestParsePageKeys(t *testing.T) {
	keys := parseInput([]byte{0x1b, '[', '5', '~'})
	if len(keys) != 1 || keys[0].Type != PageUp {
		t.Fatalf("expected PageUp, got %v", keys)
	}
}
