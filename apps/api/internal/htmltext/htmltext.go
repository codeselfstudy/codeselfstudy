// Package htmltext converts HTML email bodies to plain text suitable for LLM
// deal extraction. It is intentionally small and framework-free.
//
// Unlike a generic html-to-text library it renders anchors as "text (href)" so
// that deal URLs survive into the text the model sees, drops <style>/<script>/
// <head>, maps block elements and <br> to line breaks, and collapses runs of
// whitespace.
package htmltext

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var blockElements = map[atom.Atom]bool{
	atom.P: true, atom.Div: true, atom.Br: true, atom.Hr: true,
	atom.H1: true, atom.H2: true, atom.H3: true, atom.H4: true, atom.H5: true, atom.H6: true,
	atom.Ul: true, atom.Ol: true, atom.Li: true,
	atom.Table: true, atom.Tr: true, atom.Td: true, atom.Th: true,
	atom.Section: true, atom.Article: true, atom.Header: true, atom.Footer: true,
	atom.Nav: true, atom.Aside: true, atom.Blockquote: true,
}

var skipElements = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Head: true,
	atom.Noscript: true, atom.Title: true,
}

// Convert renders HTML source as plain text. It never returns an error for
// malformed HTML — html.Parse is tolerant — but the signature keeps an error for
// forward compatibility and reader failures.
func Convert(htmlSource string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlSource))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	walk(doc, &b)
	return collapse(b.String()), nil
}

func walk(n *html.Node, b *strings.Builder) {
	switch n.Type {
	case html.CommentNode:
		return
	case html.TextNode:
		// Collapse text-node whitespace (including source newlines) to single
		// spaces; only block elements and <br> produce actual line breaks.
		b.WriteString(spacify(n.Data))
		return
	case html.ElementNode:
		if skipElements[n.DataAtom] {
			return
		}
		if n.DataAtom == atom.A {
			writeAnchor(n, b)
			return
		}
		if blockElements[n.DataAtom] {
			b.WriteByte('\n')
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, b)
	}
	if n.Type == html.ElementNode && blockElements[n.DataAtom] {
		b.WriteByte('\n')
	}
}

// writeAnchor renders <a> as "text (href)" when the href is an http(s) URL and
// differs from the link text. A link with no text but an http href emits the
// bare URL so it is not lost (common for image buttons).
func writeAnchor(n *html.Node, b *strings.Builder) {
	var inner strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, &inner)
	}
	text := collapseInline(inner.String())

	href := ""
	for _, a := range n.Attr {
		if a.Key == "href" {
			href = strings.TrimSpace(a.Val)
			break
		}
	}
	lowerHref := strings.ToLower(href)
	isHTTP := strings.HasPrefix(lowerHref, "http://") || strings.HasPrefix(lowerHref, "https://")

	b.WriteByte(' ')
	switch {
	case text == "" && isHTTP:
		b.WriteString(href)
	case isHTTP && href != text:
		b.WriteString(text)
		b.WriteString(" (")
		b.WriteString(href)
		b.WriteByte(')')
	default:
		b.WriteString(text)
	}
	b.WriteByte(' ')
}

// spacify collapses whitespace runs in a text node to single spaces. A single
// leading or trailing space is preserved when the source had one, so words in
// adjacent inline elements stay separated across text-node boundaries. Any
// leading space that ends up at the very start of a line is removed later by
// collapse.
func spacify(s string) string {
	joined := strings.Join(strings.Fields(s), " ")
	if joined == "" {
		if s == "" {
			return ""
		}
		// A whitespace-only text node (common between inline tags in source HTML)
		// must still separate its neighbors, or "<b>50%</b> <i>off</i>" merges to
		// "50%off". collapse() trims any space that lands at a line edge.
		return " "
	}
	if strings.TrimLeftFunc(s, unicode.IsSpace) != s {
		joined = " " + joined
	}
	if strings.TrimRightFunc(s, unicode.IsSpace) != s {
		joined += " "
	}
	return joined
}

// collapseInline collapses all whitespace (including newlines) to single spaces.
func collapseInline(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// collapse trims each line, collapses intra-line whitespace runs, and drops empty
// lines, yielding one logical block per line.
func collapse(s string) string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if j := strings.Join(strings.Fields(line), " "); j != "" {
			out = append(out, j)
		}
	}
	return strings.Join(out, "\n")
}
