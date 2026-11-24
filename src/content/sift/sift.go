// Package sift extracts selected text from an article's HTML.
package sift

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type Rule struct {
	include cascadia.SelectorGroup
	exclude cascadia.SelectorGroup
}

// Compile validates the required content selector and optional exclusion selector.
func Compile(contentSelector, excludeSelector string) (*Rule, error) {
	include, err := cascadia.ParseGroup(strings.TrimSpace(contentSelector))
	if err != nil {
		return nil, fmt.Errorf("content selector: %w", err)
	}
	rule := &Rule{include: include}
	if excludeSelector = strings.TrimSpace(excludeSelector); excludeSelector != "" {
		rule.exclude, err = cascadia.ParseGroup(excludeSelector)
		if err != nil {
			return nil, fmt.Errorf("exclude selector: %w", err)
		}
	}
	return rule, nil
}

// Extract returns HTML-escaped text, with generated line breaks for the reader.
// The caller supplies the page as UTF-8.
func (r *Rule) Extract(page io.Reader) (string, error) {
	doc, err := html.Parse(page)
	if err != nil {
		return "", err
	}
	var text articleText
	r.walk(doc, &text, false, false)
	text.finishLine()
	if len(text.lines) == 0 {
		return "", errors.New("selector matched no readable text")
	}
	return strings.Join(text.lines, "<br>\n"), nil
}

type articleText struct {
	line  strings.Builder
	lines []string
}

func (t *articleText) finishLine() {
	if line := strings.Join(strings.Fields(t.line.String()), " "); line != "" {
		t.lines = append(t.lines, html.EscapeString(line))
	}
	t.line.Reset()
}

func (t *articleText) write(data string, preformatted bool) {
	if !preformatted {
		t.line.WriteString(data)
		return
	}
	for line := range strings.SplitAfterSeq(data, "\n") {
		t.line.WriteString(line)
		if strings.HasSuffix(line, "\n") {
			t.finishLine()
		}
	}
}

func (r *Rule) walk(n *html.Node, text *articleText, selected, preformatted bool) {
	switch n.Type {
	case html.TextNode:
		if selected {
			text.write(n.Data, preformatted)
		}
		return
	case html.ElementNode:
		if skipElement(n) || r.exclude.Match(n) {
			return
		}
		start := !selected && r.include.Match(n)
		selected = selected || start
		boundary := start || (selected && isBlock(n.Data))
		if boundary {
			text.finishLine()
			defer text.finishLine()
		}
		preformatted = preformatted || n.Data == "pre"
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		r.walk(child, text, selected, preformatted)
	}
}

func skipElement(n *html.Node) bool {
	switch n.Data {
	case "head", "script", "style", "noscript", "template", "iframe", "object", "canvas", "svg":
		return true
	}
	for _, attr := range n.Attr {
		if attr.Key == "hidden" || (attr.Key == "aria-hidden" && strings.EqualFold(attr.Val, "true")) {
			return true
		}
	}
	return false
}

func isBlock(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "br", "caption", "dd", "details",
		"div", "dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
		"h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li", "main", "nav",
		"ol", "p", "pre", "section", "summary", "table", "td", "th", "tr", "ul":
		return true
	}
	return false
}
