# tooey

A terminal UI runtime for Go, inspired by React and Elm.

Build full-screen terminal applications by composing declarative node trees. Tooey handles layout, diffing, and rendering — you just describe what the screen should look like.

```
go get github.com/stukennedy/tooey
```

## How it works

```
Input → Update → View → Layout → Paint → Diff → ANSI
```

Your app is three functions:

- **Init** — return your initial model (state)
- **Update** — take a message, return a new model (+ optional async commands)
- **View** — take the model, return a node tree

Tooey runs the loop at ~30fps: collect input events, call Update for each, call View once, diff the cell buffer against the previous frame, and emit only the ANSI escape sequences that changed.

## Quick example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/stukennedy/tooey/app"
    "github.com/stukennedy/tooey/input"
    "github.com/stukennedy/tooey/node"
    "golang.org/x/term"
)

type model struct {
    items    []string
    selected int
}

func main() {
    oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    a := &app.App{
        Init: func() interface{} {
            return &model{
                items:    []string{"Alpha", "Beta", "Gamma"},
                selected: 0,
            }
        },
        Update: func(m interface{}, msg app.Msg) app.UpdateResult {
            mdl := m.(*model)
            if km, ok := msg.(app.KeyMsg); ok {
                switch km.Key.Type {
                case input.Up:
                    if mdl.selected > 0 { mdl.selected-- }
                case input.Down:
                    if mdl.selected < len(mdl.items)-1 { mdl.selected++ }
                case input.RuneKey:
                    if km.Key.Rune == 'q' { return app.UpdateResult{Model: nil} }
                }
            }
            return app.NoCmd(mdl)
        },
        View: func(m interface{}, focused string) node.Node {
            mdl := m.(*model)
            items := make([]node.Node, len(mdl.items))
            for i, item := range mdl.items {
                if i == mdl.selected {
                    items[i] = node.TextStyled("> "+item, 0, 6, node.Bold)
                } else {
                    items[i] = node.Text("  "+item)
                }
            }
            return node.Column(
                node.TextStyled(" tooey ", 0, 2, node.Bold),
                node.Text(""),
                node.Box(node.BorderRounded, node.Column(items...)),
                node.Spacer(),
                node.TextStyled(" ↑/↓ navigate • q quit ", 8, 0, 0),
            )
        },
    }

    a.Run(context.Background())
}
```

## Node tree

Build your UI with value structs, not interfaces:

```go
node.Text("hello")                                    // plain text
node.TextStyled("bold red", 1, 0, node.Bold)          // styled (fg, bg, flags)
node.Row(left, right)                                  // horizontal layout
node.Column(top, middle, bottom)                       // vertical layout
node.Box(node.BorderRounded, child)                    // bordered container
node.Spacer()                                          // flex filler

// Chaining modifiers
node.Column(items...).WithFlex(1).WithScrollToBottom()
node.Text("ok").WithKey("btn").WithFocusable()
```

**Styles:** `Bold`, `Dim`, `Italic`, `Underline`, `Reverse`
**Borders:** `BorderNone`, `BorderSingle`, `BorderDouble`, `BorderRounded`
**Colors:** ANSI 256 palette (`node.Color(0)` through `node.Color(255)`, `0` = default)

## Async commands

Update can return commands — functions that run in a goroutine and send a message back:

```go
func update(m interface{}, msg app.Msg) app.UpdateResult {
    mdl := m.(*myModel)
    switch msg.(type) {
    case fetchMsg:
        mdl.loading = true
        return app.WithCmd(mdl, func() app.Msg {
            resp, _ := http.Get("https://example.com")
            defer resp.Body.Close()
            body, _ := io.ReadAll(resp.Body)
            return dataMsg{body: body}
        })
    case dataMsg:
        mdl.loading = false
        mdl.data = msg.(dataMsg).body
    }
    return app.NoCmd(mdl)
}
```

Return `app.UpdateResult{Model: nil}` to quit.

## Built-in messages

| Message | Trigger |
|---|---|
| `app.KeyMsg` | Keyboard input (wraps `input.Key`) |
| `app.ResizeMsg` | Terminal resize (SIGWINCH) |
| `app.FocusMsg` | Terminal focus gained/lost |
| `app.ScrollMsg` | Mouse scroll wheel |

## Components

The `component` package provides stateful, reusable building blocks:

- **`TextInput`** — Multi-line text input with cursor navigation, word wrap, Home/End/Up/Down support. Call `.Update(key)` in your Update function, `.Render(prefix, fg, bg)` in View.
- **`List`** — Vertical selection list with highlight styling.
- **`TextBlock`** — Styled text span with optional key.
- **`Box`** — Bordered container with title.

## Focus management

Tooey manages focus automatically. Mark nodes as focusable, give them a key, and the framework handles Tab/Shift-Tab cycling and Escape to pop context:

```go
node.Text("clickable").WithKey("btn-1").WithFocusable()
```

The `focused` string passed to your View function is the key of the currently focused node.

## Scrolling

Columns, Lists, and Panes support vertical scrolling:

```go
node.Column(children...).WithScrollOffset(offset)     // manual scroll position
node.Column(children...).WithScrollToBottom()          // auto-scroll to end
```

Combine both for chat-style UIs where new content auto-scrolls but the user can scroll up.

## Server-driven UI (SSE)

The `sse` package connects your TUI to a server. The client auto-reconnects and feeds events into your Update loop as messages:

```go
client := &sse.Client{URL: "http://localhost:8080/events"}
ch, _ := client.Connect(ctx)

// In your app loop, read from ch and dispatch as app messages

// Send actions back to the server:
sse.PostAction("http://localhost:8080/action", "submit", payload)
```

## Render pipeline internals

Each frame passes through five stages:

1. **View** — Your function builds a `node.Node` tree (immutable value structs)
2. **Layout** — Single-pass flex engine computes a `layout.LayoutNode` tree with absolute `(x, y, w, h)` positions
3. **Paint** — Walks the layout tree, writes runes + styles into a flat `cell.Buffer` (row-major `[]Cell`)
4. **Diff** — Compares current buffer against previous frame, groups adjacent changed cells into horizontal runs
5. **Render** — Emits minimal ANSI escape sequences (cursor moves + SGR attributes) for only the changed runs

The buffer is `width × height` cells. Each `Cell` holds a rune, foreground color, background color, and style flags. Diffing is a single linear scan — O(width × height) with early exit on unchanged rows.

## Demos

```bash
go run ./demos/list      # Interactive list with selection and activation counter
go run ./demos/maude     # Chat TUI with tool blocks, markdown rendering, and diff highlighting
```

## Requirements

- Go 1.24+
- A terminal that supports ANSI escape sequences (most modern terminals)
- Only external dependency: `golang.org/x/term`

## License

MIT
