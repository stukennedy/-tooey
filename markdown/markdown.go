package markdown

import (
	"strings"

	"github.com/stukennedy/tooey/node"
)

// ColorScheme customizes colors for markdown elements.
type ColorScheme struct {
	Text    node.Color
	Heading node.Color
	Code    node.Color
	CodeBG  node.Color
	Quote   node.Color
	Link    node.Color
	Bullet  node.Color
	CheckOn node.Color
	CheckOff node.Color
	Rule    node.Color
}

// DefaultColors returns a sensible default color scheme.
func DefaultColors(fg node.Color) ColorScheme {
	return ColorScheme{
		Text:     fg,
		Heading:  33,  // blue
		Code:     223, // warm white
		CodeBG:   236, // dark gray
		Quote:    245, // muted gray
		Link:     39,  // cyan-blue
		Bullet:   44,  // cyan
		CheckOn:  34,  // green
		CheckOff: 245, // gray
		Rule:     240, // dim gray
	}
}

// Render converts markdown text into styled node.Node trees using default colors.
func Render(text string, width int, fg node.Color) []node.Node {
	return RenderWithColors(text, width, DefaultColors(fg))
}

// RenderWithColors converts markdown text into styled node.Node trees.
func RenderWithColors(text string, width int, colors ColorScheme) []node.Node {
	lines := strings.Split(text, "\n")
	var nodes []node.Node
	i := 0
	for i < len(lines) {
		line := lines[i]

		// Code fence
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			i++
			var codeLines []node.Node
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				codeLines = append(codeLines, node.TextStyled(lines[i], colors.Code, colors.CodeBG, 0))
				i++
			}
			if i < len(lines) {
				i++ // skip closing fence
			}
			if len(codeLines) == 0 {
				codeLines = append(codeLines, node.TextStyled(" ", colors.Code, colors.CodeBG, 0))
			}
			nodes = append(nodes, node.Box(node.BorderRounded, node.Column(codeLines...)))
			continue
		}

		// Horizontal rule
		trimmed := strings.TrimSpace(line)
		if isHorizontalRule(trimmed) {
			if width <= 0 {
				width = 40
			}
			nodes = append(nodes, node.TextStyled(strings.Repeat("─", width), colors.Rule, 0, 0))
			i++
			continue
		}

		// Heading
		if level, content := parseHeading(line); level > 0 {
			spans := parseInline(content, colors.Heading, colors)
			for idx := range spans {
				spans[idx].Props.Style |= node.Bold
			}
			nodes = append(nodes, node.Row(spans...))
			i++
			continue
		}

		// Checkbox list items
		if rest, checked := parseCheckbox(line); checked >= 0 {
			if checked == 1 {
				prefix := node.TextStyled("  ✔ ", colors.CheckOn, 0, 0)
				spans := parseInline(rest, colors.Text, colors)
				nodes = append(nodes, node.Row(append([]node.Node{prefix}, spans...)...))
			} else {
				prefix := node.TextStyled("  ☐ ", colors.CheckOff, 0, 0)
				spans := parseInline(rest, colors.Text, colors)
				for idx := range spans {
					spans[idx].Props.Style |= node.Dim
				}
				nodes = append(nodes, node.Row(append([]node.Node{prefix}, spans...)...))
			}
			i++
			continue
		}

		// Bullet list
		if rest, ok := parseBullet(line); ok {
			prefix := node.TextStyled("  • ", colors.Bullet, 0, 0)
			spans := parseInline(rest, colors.Text, colors)
			nodes = append(nodes, node.Row(append([]node.Node{prefix}, spans...)...))
			i++
			continue
		}

		// Numbered list
		if num, rest, ok := parseNumbered(line); ok {
			prefix := node.TextStyled("  "+num+". ", colors.Bullet, 0, 0)
			spans := parseInline(rest, colors.Text, colors)
			nodes = append(nodes, node.Row(append([]node.Node{prefix}, spans...)...))
			i++
			continue
		}

		// Blockquote
		if rest, ok := parseBlockquote(line); ok {
			prefix := node.TextStyled("  │ ", colors.Quote, 0, 0)
			spans := parseInline(rest, colors.Text, colors)
			for idx := range spans {
				spans[idx].Props.Style |= node.Italic
			}
			nodes = append(nodes, node.Row(append([]node.Node{prefix}, spans...)...))
			i++
			continue
		}

		// Empty line
		if trimmed == "" {
			nodes = append(nodes, node.Text(""))
			i++
			continue
		}

		// Plain paragraph
		spans := parseInline(line, colors.Text, colors)
		nodes = append(nodes, node.Row(spans...))
		i++
	}
	return nodes
}

func isHorizontalRule(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	if len(s) < 3 {
		return false
	}
	if strings.Count(s, "-") == len(s) || strings.Count(s, "*") == len(s) || strings.Count(s, "_") == len(s) {
		return true
	}
	return false
}

func parseHeading(line string) (int, string) {
	t := strings.TrimSpace(line)
	level := 0
	for _, c := range t {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	if level > 0 && level <= 6 && len(t) > level && t[level] == ' ' {
		return level, strings.TrimSpace(t[level+1:])
	}
	return 0, ""
}

func parseCheckbox(line string) (string, int) {
	t := strings.TrimSpace(line)
	for _, prefix := range []string{"- ", "* "} {
		if strings.HasPrefix(t, prefix) {
			rest := t[len(prefix):]
			if strings.HasPrefix(rest, "[x] ") || strings.HasPrefix(rest, "[X] ") {
				return rest[4:], 1
			}
			if strings.HasPrefix(rest, "[ ] ") {
				return rest[4:], 0
			}
		}
	}
	return "", -1
}

func parseBullet(line string) (string, bool) {
	t := strings.TrimSpace(line)
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(t, prefix) {
			return t[len(prefix):], true
		}
	}
	return "", false
}

func parseNumbered(line string) (string, string, bool) {
	t := strings.TrimSpace(line)
	for i, c := range t {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' && i > 0 && i+1 < len(t) && t[i+1] == ' ' {
			return t[:i], t[i+2:], true
		}
		break
	}
	return "", "", false
}

func parseBlockquote(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "> ") {
		return t[2:], true
	}
	if t == ">" {
		return "", true
	}
	return "", false
}

// parseInline scans text for inline markdown and returns styled nodes.
func parseInline(text string, defaultFG node.Color, colors ColorScheme) []node.Node {
	var nodes []node.Node
	var buf strings.Builder
	runes := []rune(text)
	i := 0

	flush := func(fg node.Color, style node.StyleFlags) {
		if buf.Len() > 0 {
			nodes = append(nodes, node.TextStyled(buf.String(), fg, 0, style))
			buf.Reset()
		}
	}

	for i < len(runes) {
		// Bold+Italic ***
		if i+2 < len(runes) && runes[i] == '*' && runes[i+1] == '*' && runes[i+2] == '*' {
			if end := findClose(runes, i+3, "***"); end >= 0 {
				flush(defaultFG, 0)
				nodes = append(nodes, node.TextStyled(string(runes[i+3:end]), defaultFG, 0, node.Bold|node.Italic))
				i = end + 3
				continue
			}
		}
		// Bold **
		if i+1 < len(runes) && runes[i] == '*' && runes[i+1] == '*' {
			if end := findClose(runes, i+2, "**"); end >= 0 {
				flush(defaultFG, 0)
				nodes = append(nodes, node.TextStyled(string(runes[i+2:end]), defaultFG, 0, node.Bold))
				i = end + 2
				continue
			}
			// No closing **, treat as literal
			buf.WriteRune(runes[i])
			buf.WriteRune(runes[i+1])
			i += 2
			continue
		}
		// Italic *
		if runes[i] == '*' {
			if end := findClose(runes, i+1, "*"); end >= 0 && end > i+1 {
				flush(defaultFG, 0)
				nodes = append(nodes, node.TextStyled(string(runes[i+1:end]), defaultFG, 0, node.Italic))
				i = end + 1
				continue
			}
		}
		// Inline code `
		if runes[i] == '`' {
			if end := findClose(runes, i+1, "`"); end >= 0 {
				flush(defaultFG, 0)
				nodes = append(nodes, node.TextStyled(string(runes[i+1:end]), colors.Code, colors.CodeBG, 0))
				i = end + 1
				continue
			}
		}
		// Link [text](url)
		if runes[i] == '[' {
			if linkText, end := parseLink(runes, i); end >= 0 {
				flush(defaultFG, 0)
				nodes = append(nodes, node.TextStyled(linkText, colors.Link, 0, node.Underline))
				i = end
				continue
			}
		}
		buf.WriteRune(runes[i])
		i++
	}
	flush(defaultFG, 0)
	return nodes
}

func findClose(runes []rune, start int, marker string) int {
	mr := []rune(marker)
	for i := start; i <= len(runes)-len(mr); i++ {
		match := true
		for j := 0; j < len(mr); j++ {
			if runes[i+j] != mr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func parseLink(runes []rune, start int) (string, int) {
	// [text](url)
	closeB := -1
	for i := start + 1; i < len(runes); i++ {
		if runes[i] == ']' {
			closeB = i
			break
		}
	}
	if closeB < 0 || closeB+1 >= len(runes) || runes[closeB+1] != '(' {
		return "", -1
	}
	closeP := -1
	for i := closeB + 2; i < len(runes); i++ {
		if runes[i] == ')' {
			closeP = i
			break
		}
	}
	if closeP < 0 {
		return "", -1
	}
	return string(runes[start+1 : closeB]), closeP + 1
}
