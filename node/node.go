package node

// NodeType identifies the kind of UI node.
type NodeType int

const (
	TextNode NodeType = iota
	BoxNode
	RowNode
	ColumnNode
	ListNode
	PaneNode
	SpacerNode
)

// Color represents an ANSI 256-color value. 0 means default/unset.
type Color uint8

// StyleFlags are bitwise text style attributes.
type StyleFlags uint8

const (
	Bold      StyleFlags = 1 << iota
	Dim
	Italic
	Underline
	Reverse
)

// BorderStyle defines box border appearance.
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderDouble
	BorderRounded
)

// Props holds configurable properties for a node.
type Props struct {
	Text       string
	Width      int // 0 = auto
	Height     int // 0 = auto
	FlexWeight int // 0 = no flex, >0 = relative weight
	Border     BorderStyle
	Focusable  bool
	Key        string
	FG           Color
	BG           Color
	Style        StyleFlags
	ScrollOffset   int  // vertical scroll offset for Column/List/Pane
	ScrollToBottom bool // auto-scroll so bottom content is visible
}

// Node represents a virtual UI element in the component tree.
type Node struct {
	Type     NodeType
	Props    Props
	Children []Node
}

// Builder functions

func Text(s string) Node {
	return Node{Type: TextNode, Props: Props{Text: s}}
}

func TextStyled(s string, fg, bg Color, style StyleFlags) Node {
	return Node{Type: TextNode, Props: Props{Text: s, FG: fg, BG: bg, Style: style}}
}

func Row(children ...Node) Node {
	return Node{Type: RowNode, Children: children}
}

func Column(children ...Node) Node {
	return Node{Type: ColumnNode, Children: children}
}

func Box(border BorderStyle, child Node) Node {
	return Node{Type: BoxNode, Props: Props{Border: border}, Children: []Node{child}}
}

func List(children ...Node) Node {
	return Node{Type: ListNode, Children: children}
}

func Pane(children ...Node) Node {
	return Node{Type: PaneNode, Children: children}
}

func Spacer() Node {
	return Node{Type: SpacerNode, Props: Props{FlexWeight: 1}}
}

// WithKey sets the key on a node and returns it.
func (n Node) WithKey(key string) Node {
	n.Props.Key = key
	return n
}

// WithFlex sets the flex weight and returns the node.
func (n Node) WithFlex(weight int) Node {
	n.Props.FlexWeight = weight
	return n
}

// WithSize sets explicit width/height and returns the node.
func (n Node) WithSize(w, h int) Node {
	n.Props.Width = w
	n.Props.Height = h
	return n
}

// WithFocusable marks the node as focusable.
func (n Node) WithFocusable() Node {
	n.Props.Focusable = true
	return n
}

// WithScrollOffset sets the vertical scroll offset.
func (n Node) WithScrollOffset(offset int) Node {
	n.Props.ScrollOffset = offset
	return n
}

func (n Node) WithScrollToBottom() Node {
	n.Props.ScrollToBottom = true
	return n
}
