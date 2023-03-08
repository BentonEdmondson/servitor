package hypertext

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
	"regexp"
	"mimicry/style"
	"errors"
	"mimicry/ansi"
)

// TODO: create a `bulletedList` function for all situations where
// html-specific wrapping is needed. Put them at the bottom of the file
// and note that that section is different. (This will include one for
// headers)

// TODO: blocks need to be trimmed on the inside and newlined on 
// the outside

/* Terminal codes and control characters should already be escaped
   by this point */
func Render(text string, width int) (string, error) {
	nodes, err := html.ParseFragment(strings.NewReader(text), &html.Node{
		Type: html.ElementNode,
		Data: "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return "", err
	}
	rendered, err := renderList(nodes, width)
	if err != nil {
		return "", err
	}

	wrapped := ansi.Wrap(rendered, width)
	return strings.Trim(wrapped, " \n"), nil
}

func renderList(nodes []*html.Node, width int) (string, error) {
	output := ""
	for _, current := range nodes {
		result, err := renderNode(current, width, false)
		if err != nil {
			return "", err
		}
		output = mergeText(output, result)
	}
	return output, nil
}

/* 	Merges text according to the following rules:
	1. Extract trailing whitespace from lhs and
	   leading whitespace from rhs and concat them.
	2. Append the two sides in the following way,
	   depending on the extracted whitespace:
	   	- If it is empty, append the sides
		- Else, if it contains 0 newlines, append
		  the sides with a single space between.
		- Else, if it contains 1 newline, append
		  the sides with a single newline between.
		- Else, append the sides with 2 newlines
		  between.
*/
func mergeText(lhs string, rhs string) string {
	trimRight := regexp.MustCompile(`(?s)^(.*?)([ \n]*)$`)
	lhsMatches := trimRight.FindStringSubmatch(lhs)
	lhsTrimmed := lhsMatches[1]

	trimLeft := regexp.MustCompile(`(?s)^([ \n]*)(.*)$`)
	rhsMatches := trimLeft.FindStringSubmatch(rhs)
	rhsTrimmed := rhsMatches[2]

	whitespace := lhsMatches[2] + rhsMatches[1]

	if whitespace == "" {
		return lhsTrimmed + rhsTrimmed
	}

	switch strings.Count(whitespace, "\n") {
	case 0: return lhsTrimmed + " " + rhsTrimmed
	case 1: return lhsTrimmed + "\n" + rhsTrimmed
	}

	return lhsTrimmed + "\n\n" + rhsTrimmed
}

func renderNode(node *html.Node, width int, preserveWhitespace bool) (string, error) {
	if node.Type == html.TextNode {
		if !preserveWhitespace {
			whitespace := regexp.MustCompile(`[ \t\n\r]+`)
			return whitespace.ReplaceAllString(node.Data, " "), nil
		}
		return node.Data, nil
	}

	if node.Type != html.ElementNode {
		return "", nil
	}

	content, err := renderChildren(node, width, preserveWhitespace)
	if err != nil {
		return "", err
	}

	switch node.Data {
	case "a":
		return style.Link(content), nil
	case "s", "del":
		return style.Strikethrough(content), nil
	case "code":
		return style.Code(content), nil
	case "i", "em":
		return style.Italic(content), nil
	case "b", "strong":
		return style.Bold(content), nil
	case "u":
		return style.Underline(content), nil
	case "mark":
		return style.Highlight(content), nil
	case "span", "li":
		return content, nil
	case "br":
		return "\n", nil

	case "p", "div":
		return block(content), nil
	case "pre":
		content, err := renderChildren(node, width - 2, true)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width, true)
		return block(style.CodeBlock(wrapped)), err
	case "blockquote":
		content, err := renderChildren(node, width - 1, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 1, preserveWhitespace)
		// TODO: this text wrap is ugly
		return block(style.QuoteBlock(strings.Trim(wrapped, " \n"))), nil
	case "ul":
		list, err := bulletedList(node, width, preserveWhitespace)
		return list, err
	// case "ul":
	// 	return numberedList(node), nil

	case "h1":
		content, err := renderChildren(node, width - 2, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 2, preserveWhitespace)
		return block(style.Header(wrapped, 1)), nil
	case "h2":
		content, err := renderChildren(node, width - 3, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 3, preserveWhitespace)
		return block(style.Header(wrapped, 2)), nil
	case "h3":
		content, err := renderChildren(node, width - 4, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 4, preserveWhitespace)
		return block(style.Header(wrapped, 3)), nil
	case "h4":
		content, err := renderChildren(node, width - 5, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 5, preserveWhitespace)
		return block(style.Header(wrapped, 4)), nil
	case "h5":
		content, err := renderChildren(node, width - 6, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 6, preserveWhitespace)
		return block(style.Header(wrapped, 5)), nil
	case "h6":
		content, err := renderChildren(node, width - 7, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(content, width - 7, preserveWhitespace)
		return block(style.Header(wrapped, 6)), nil

	case "hr":
		return block(strings.Repeat("―", width)), nil
	case "img", "video", "audio", "iframe":
		text := getAttribute("alt", node.Attr)
		if text == "" {
			text = getAttribute("title", node.Attr)
		}
		if text == "" {
			text = getAttribute("src", node.Attr)
		}
		if text == "" {
			return "", errors.New(node.Data + " tag is missing both `alt` and `src` attributes")
		}
		wrapped := situationalWrap(text, width - 2, preserveWhitespace)
		return block(style.LinkBlock(wrapped)), nil
	}

	return "", errors.New("Encountered unrecognized element " + node.Data)
}

func renderChildren(node *html.Node, width int, preserveWhitespace bool) (string, error) {
	output := ""
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		result, err := renderNode(current, width, preserveWhitespace)
		if err != nil {
			return "", err
		}
		output = mergeText(output, result)
	}
	return output, nil
}

func block(text string) string {
	return "\n\n" + strings.Trim(text, " \n") + "\n\n"
}

func bulletedList(node *html.Node, width int, preserveWhitespace bool) (string, error) {
	output := ""
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		if current.Type != html.ElementNode {
			continue
		}

		if current.Data != "li" {
			continue
		}

		result, err := renderNode(current, width - 2, preserveWhitespace)
		if err != nil {
			return "", err
		}
		wrapped := situationalWrap(result, width - 2, preserveWhitespace)
		output += "\n" + style.Bullet(wrapped)
	}

	if node.Parent == nil {
		return block(output), nil
	} else if node.Parent.Data == "li" {
		return output, nil
	} else {
		return block(output), nil
	}
}

func getAttribute(name string, attributes []html.Attribute) string {
	for _, attribute := range attributes {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func situationalWrap(text string, width int, preserveWhitespace bool) string {
	if preserveWhitespace {
		return ansi.DumbWrap(text, width)
	}

	return ansi.Wrap(text, width)
}