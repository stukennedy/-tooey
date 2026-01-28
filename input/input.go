package input

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Ensure syscall is used for SIGWINCH
var _ = syscall.SIGWINCH

// KeyType identifies the kind of key event.
type KeyType int

const (
	RuneKey KeyType = iota
	Up
	Down
	Left
	Right
	Tab
	ShiftTab
	Enter
	Escape
	Backspace
	Delete
	Home
	End
	PageUp
	PageDown
	CtrlC
	CtrlD
	CtrlZ
	ShiftEnter
	FocusIn
	FocusOut
	MouseClick
	MouseScrollUp
	MouseScrollDown
)

// Key represents a keyboard input event.
type Key struct {
	Type KeyType
	Rune rune
}

// ResizeMsg indicates the terminal was resized.
type ResizeMsg struct {
	Width, Height int
}

// ReadKeys reads raw terminal input and sends parsed Key events.
func ReadKeys(ctx context.Context, r io.Reader) <-chan Key {
	ch := make(chan Key, 32)
	go func() {
		defer close(ch)
		buf := make([]byte, 64)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := r.Read(buf)
			if err != nil {
				return
			}

			keys := parseInput(buf[:n])
			for _, k := range keys {
				select {
				case ch <- k:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}

// WatchResize listens for SIGWINCH and sends ResizeMsg events.
func WatchResize(ctx context.Context) <-chan ResizeMsg {
	ch := make(chan ResizeMsg, 4)
	sigCh := make(chan os.Signal, 4)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		defer close(ch)
		defer signal.Stop(sigCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				w, h := TermSize()
				select {
				case ch <- ResizeMsg{Width: w, Height: h}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}

// TermSize returns the current terminal width and height.
func TermSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

// parseInput parses raw bytes into Key events.
func parseInput(data []byte) []Key {
	var keys []Key
	i := 0
	for i < len(data) {
		if data[i] == 0x1b { // ESC
			if i+1 < len(data) && data[i+1] == '[' {
				// CSI sequence
				k, consumed := parseCSI(data[i+2:])
				if consumed > 0 {
					keys = append(keys, k)
					i += 2 + consumed
					continue
				}
			}
			// Alt+Enter (ESC followed by CR or LF) → ShiftEnter
			if i+1 < len(data) && (data[i+1] == '\r' || data[i+1] == '\n') {
				keys = append(keys, Key{Type: ShiftEnter})
				i += 2
				continue
			}
			keys = append(keys, Key{Type: Escape})
			i++
		} else if data[i] == '\r' {
			keys = append(keys, Key{Type: Enter})
			i++
		} else if data[i] == '\n' {
			// Ctrl+J or literal newline → newline insertion
			keys = append(keys, Key{Type: ShiftEnter})
			i++
		} else if data[i] == '\t' {
			keys = append(keys, Key{Type: Tab})
			i++
		} else if data[i] == 0x7f || data[i] == '\b' {
			keys = append(keys, Key{Type: Backspace})
			i++
		} else if data[i] == 0x03 { // Ctrl+C
			keys = append(keys, Key{Type: CtrlC})
			i++
		} else if data[i] == 0x04 { // Ctrl+D
			keys = append(keys, Key{Type: CtrlD})
			i++
		} else if data[i] == 0x1a { // Ctrl+Z
			keys = append(keys, Key{Type: CtrlZ})
			i++
		} else if data[i] >= 0x20 { // printable or multi-byte UTF-8
			r, size := decodeRune(data[i:])
			keys = append(keys, Key{Type: RuneKey, Rune: r})
			i += size
		} else {
			i++ // skip unknown control chars
		}
	}
	return keys
}

func parseCSI(data []byte) (Key, int) {
	if len(data) == 0 {
		return Key{}, 0
	}
	switch data[0] {
	case 'A':
		return Key{Type: Up}, 1
	case 'B':
		return Key{Type: Down}, 1
	case 'C':
		return Key{Type: Right}, 1
	case 'D':
		return Key{Type: Left}, 1
	case 'H':
		return Key{Type: Home}, 1
	case 'F':
		return Key{Type: End}, 1
	case 'Z':
		return Key{Type: ShiftTab}, 1
	case 'I':
		return Key{Type: FocusIn}, 1
	case 'O':
		return Key{Type: FocusOut}, 1
	}
	// Handle sequences like \x1b[5~ (PageUp), \x1b[6~ (PageDown), \x1b[3~ (Delete)
	if len(data) >= 2 && data[1] == '~' {
		switch data[0] {
		case '3':
			return Key{Type: Delete}, 2
		case '5':
			return Key{Type: PageUp}, 2
		case '6':
			return Key{Type: PageDown}, 2
		}
	}
	// Kitty keyboard protocol: \x1b[13;2u = Shift+Enter
	if len(data) >= 4 && data[0] == '1' && data[1] == '3' && data[2] == ';' && data[3] == '2' {
		if len(data) >= 5 && data[4] == 'u' {
			return Key{Type: ShiftEnter}, 5
		}
	}
	// SGR mouse: \x1b[<btn;x;yM or \x1b[<btn;x;ym
	if len(data) >= 1 && data[0] == '<' {
		for j := 1; j < len(data); j++ {
			if data[j] == 'M' || data[j] == 'm' {
				// Parse button number from data[1:j] (e.g. "64;10;20")
				btn := parseSGRButton(data[1:j])
				kt := MouseClick
				switch btn {
				case 64:
					kt = MouseScrollUp
				case 65:
					kt = MouseScrollDown
				}
				return Key{Type: kt}, j + 1
			}
		}
	}
	// Normal mouse: \x1b[M + 3 bytes (btn, x, y)
	if len(data) >= 1 && data[0] == 'M' && len(data) >= 4 {
		btn := data[1] - 32
		kt := MouseClick
		switch btn {
		case 64:
			kt = MouseScrollUp
		case 65:
			kt = MouseScrollDown
		}
		return Key{Type: kt}, 4
	}
	return Key{}, 0
}

// parseSGRButton extracts the button number from SGR mouse data like "64;10;20".
func parseSGRButton(data []byte) int {
	n := 0
	for _, b := range data {
		if b == ';' {
			break
		}
		if b >= '0' && b <= '9' {
			n = n*10 + int(b-'0')
		}
	}
	return n
}

func decodeRune(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0, 0
	}
	b := data[0]
	if b < 0x80 {
		return rune(b), 1
	}
	// Simple UTF-8 decode
	var r rune
	var size int
	switch {
	case b&0xE0 == 0xC0:
		size = 2
		r = rune(b & 0x1F)
	case b&0xF0 == 0xE0:
		size = 3
		r = rune(b & 0x0F)
	case b&0xF8 == 0xF0:
		size = 4
		r = rune(b & 0x07)
	default:
		return '?', 1
	}
	if len(data) < size {
		return '?', 1
	}
	for i := 1; i < size; i++ {
		r = r<<6 | rune(data[i]&0x3F)
	}
	return r, size
}
