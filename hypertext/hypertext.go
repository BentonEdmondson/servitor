package hypertext

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"mimicry/ansi"
	"mimicry/style"
	"regexp"
	"strings"
)

type Markup struct {
	tree        []*html.Node
	cached      string
	cachedWidth int
}

type context struct {
	preserveWhitespace bool
	width              int
	links              *[]string
}

func NewMarkup(text string) (*Markup, []string, error) {
	nodes, err := html.ParseFragment(strings.NewReader(text), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return nil, []string{}, err
	}
	rendered, links := renderWithLinks(nodes, 80)
	return &Markup{
		tree:        nodes,
		cached:      rendered,
		cachedWidth: 80,
	}, links, nil
}

func (m *Markup) Render(width int) string {
	if m.cachedWidth == width {
		return m.cached
	}
	rendered, _ := renderWithLinks(m.tree, width)
	m.cachedWidth = width
	m.cached = rendered
	return rendered
}

func renderWithLinks(nodes []*html.Node, width int) (string, []string) {
	ctx := context{
		preserveWhitespace: false,
		width:              width,
		links:              &[]string{},
	}
	output := ""
	for _, current := range nodes {
		result := renderNode(current, ctx)
		output = mergeText(output, result)
	}
	output = ansi.Wrap(output, width)
	return strings.Trim(output, " \n"), *ctx.links
}

/*
		Merges text according to the following rules:
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

	newlineCount := strings.Count(whitespace, "\n")

	if newlineCount == 0 {
		return lhsTrimmed + " " + rhsTrimmed
	}

	if newlineCount == 1 {
		return lhsTrimmed + "\n" + rhsTrimmed
	}

	return lhsTrimmed + "\n\n" + rhsTrimmed
}

func renderNode(node *html.Node, ctx context) string {
	if node.Type == html.TextNode {
		if !ctx.preserveWhitespace {
			whitespace := regexp.MustCompile(`[ \t\n\r]+`)
			return whitespace.ReplaceAllString(node.Data, " ")
		}
		return node.Data
	}

	if node.Type != html.ElementNode {
		return ""
	}

	switch node.Data {
	case "a":
		link := getAttribute("href", node.Attr)
		if link == "" {
			return renderChildren(node, ctx)
		}
		*ctx.links = append(*ctx.links, link)
		/* This must occur before the styling because it mutates ctx.links */
		rendered := renderChildren(node, ctx)
		return style.Link(rendered, len(*ctx.links))
	case "s", "del":
		return style.Strikethrough(renderChildren(node, ctx))
	case "code":
		ctx.preserveWhitespace = true
		return style.Code(renderChildren(node, ctx))
	case "i", "em":
		return style.Italic(renderChildren(node, ctx))
	case "b", "strong":
		return style.Bold(renderChildren(node, ctx))
	case "u", "ins":
		return style.Underline(renderChildren(node, ctx))
	case "mark":
		return style.Highlight(renderChildren(node, ctx))
	case "span":
		return renderChildren(node, ctx)
	case "li":
		return strings.Trim(renderChildren(node, ctx), " \n")
	case "br":
		return "\n"

	case "p", "div":
		return block(renderChildren(node, ctx))
	case "pre":
		ctx.preserveWhitespace = true
		wrapped := ansi.Pad(situationalWrap(renderChildren(node, ctx), ctx), ctx.width)
		return block(style.CodeBlock(wrapped))
	case "blockquote":
		ctx.width -= 1
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.QuoteBlock(strings.Trim(wrapped, " \n")))
	case "ul":
		return bulletedList(node, ctx)
	// case "ul":
	// 	return numberedList(node)
	case "h1":
		ctx.width -= 2
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 1))
	case "h2":
		ctx.width -= 3
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 2))
	case "h3":
		ctx.width -= 4
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 3))
	case "h4":
		ctx.width -= 5
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 4))
	case "h5":
		ctx.width -= 6
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 5))
	case "h6":
		ctx.width -= 7
		wrapped := situationalWrap(renderChildren(node, ctx), ctx)
		return block(style.Header(wrapped, 6))
	case "hr":
		return block(strings.Repeat("\u23AF", ctx.width))

	/*
		The spec does not define the alt attribute for videos nor audio.
		I think it should, so if present I display it. It is
		tempting to use the children of the video and audio tags for
		this purpose, but it looks like they exist more so for backwards
		compatibility, so should contain something like "your browser does
		not support inline video; click here" as opposed to actual alt
		text.
	*/
	case "img", "video", "audio":
		alt := getAttribute("alt", node.Attr)
		link := getAttribute("src", node.Attr)
		if alt == "" {
			alt = link
		}
		if link == "" {
			return block(alt)
		}
		*ctx.links = append(*ctx.links, link)
		ctx.width -= 2
		wrapped := situationalWrap(alt, ctx)
		return block(style.LinkBlock(wrapped, len(*ctx.links)))
	case "iframe":
		alt := getAttribute("title", node.Attr)
		link := getAttribute("src", node.Attr)
		if alt == "" {
			alt = link
		}
		if link == "" {
			return block(alt)
		}
		*ctx.links = append(*ctx.links, link)
		ctx.width -= 2
		wrapped := situationalWrap(alt, ctx)
		return block(style.LinkBlock(wrapped, len(*ctx.links)))
	default:
		return bad(node, ctx)
	}
}

func renderChildren(node *html.Node, ctx context) string {
	output := ""
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		result := renderNode(current, ctx)
		output = mergeText(output, result)
	}
	return output
}

func block(text string) string {
	return "\n\n" + strings.Trim(text, " \n") + "\n\n"
}

func bulletedList(node *html.Node, ctx context) string {
	output := ""
	ctx.width -= 2
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		if current.Type != html.ElementNode {
			continue
		}

		result := ""
		if current.Data != "li" {
			result = bad(current, ctx)
		} else {
			result = renderNode(current, ctx)
		}

		wrapped := situationalWrap(result, ctx)
		output += "\n" + style.Bullet(wrapped)
	}

	if node.Parent != nil && node.Parent.Data == "li" {
		return output
	}
	return block(output)
}

func bad(node *html.Node, ctx context) string {
	return style.Red("<"+node.Data+">") + renderChildren(node, ctx) + style.Red("</"+node.Data+">")
}

func getAttribute(name string, attributes []html.Attribute) string {
	for _, attribute := range attributes {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func situationalWrap(text string, ctx context) string {
	if ctx.preserveWhitespace {
		return ansi.DumbWrap(text, ctx.width)
	}

	return ansi.Wrap(text, ctx.width)
}
