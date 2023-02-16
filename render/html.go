package render

import (
	"golang.org/x/net/html"
	"fmt"
	"mimicry/style"
	"errors"
	"strings"
	"regexp"
	"golang.org/x/net/html/atom"
)

// preprocessing:
// substitute escape key for escape visualizer

// newline strategy:
// blocks have 2 newlines before and two after,
// at the end collapse 4 newlines into 2
// maybe instead collapse any amount greater
// (regex: \n{2,}) than 2 down to 2
// also at the end trim all newlines from
// very beginning and very end

// for block links probably use ‣

// I think it may work to collapse all
// text node whitespace down to space,
// and then trim the contents of blocks
// (including the implicit body element)

// FIXME: instead, you should collapse all whitespace into
// space (including newline, so format newlines don't appear), then do newline insertion for blocks
// then collapse all single-newline-containing whitespace into
// one newline and multi-newline-containing whitespace into two
// newlines

// I will have this issue: https://unix.stackexchange.com/questions/170551/force-lynx-or-elinks-to-interpret-spaces-and-line-breaks
// i.e. 3 or more br's in a row become idempotent, but I don't care

func renderHTML(markup string) (string, error) {
	/* 	Preprocessing
		To prevent input text from changing its color, style, etc
		via terminal escape codes, swap out U+001B (ESCAPE) for
		U+241B (SYMBOL FOR ESCAPE)

		TODO: move this to the complete beginning of render, not
		just the HTML section
	*/
	markup = strings.ReplaceAll(markup, "\u001b", "␛")


	nodes, err := html.ParseFragment(strings.NewReader(markup), &html.Node{
		Type: html.ElementNode,
		Data: "body",
		DataAtom: atom.Body,
	})
	if err != nil {
		return "", err
	}
	serialized, err := SerializeList(nodes)
	if err != nil {
		return "", err
	}

	/*
		Postprocessing
		Block elements are separated from siblings by prepending
		and appending two newline characters. If two blocks are
		adjacent, this will result in too many newline characters.
		Furthermore, in text nodes, newline-containing whitespace
		is collapsed into a single newline, potentially resulting
		in even more newlines. So collapse sequences of over two
		newlines into two newlines. Also trim all newlines from
		the beginning and end of the output.
	*/
	manyNewlines := regexp.MustCompile(`\n{2,}`)
	serialized = manyNewlines.ReplaceAllString(serialized, "\n\n")
	serialized = strings.Trim(serialized, "\n")
	return serialized, nil
}

func renderNode(node *html.Node, preserveWhitespace bool) (string, error) {
	if node.Type == html.TextNode {
		if !preserveWhitespace {
			whitespace := regexp.MustCompile(`[\t ]+`)
			newline := regexp.MustCompile(`[\n\t ]*\n[n\t ]*`)
			processed := newline.ReplaceAllString(node.Data, "\n")
			processed = whitespace.ReplaceAllString(processed, " ")
			return processed, nil
		}
		return node.Data, nil
	}

	if node.Type != html.ElementNode {
		return "", nil
	}

	// this may need to be moved down into the switch
	// so that pre and code can override the last parameter
	content := serializeChildren(node, preserveWhitespace)

	switch node.Data {
	case "a":
		return style.Linkify(content), nil
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
		return block(style.CodeBlock(content)), nil
	case "blockquote":
		// FIXME: change blockquote to style.QuoteBlock
		return block(blockquote(content)), nil
	case "ul":
		return block(bulletedList(node, preserveWhitespace)), nil
	// case "ul":
	// 	return numberedList(node), nil
	}

	return "", errors.New("Encountered unrecognized element " + node.Data)
}

func serializeChildren(node *html.Node, preserveWhitespace bool) (string) {
	output := ""
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		result, _ := renderNode(current, preserveWhitespace)
		// if err != nil {
		// 	return "", err
		// }
		output += result
	}
	return output
}

func SerializeList(nodes []*html.Node) (string, error) {
	output := ""
	for _, current := range nodes {
		result, err := renderNode(current, false)
		if err != nil {
			return "", err
		}
		output += result
	}
	return output, nil
}

func block(text string) string {
	return fmt.Sprintf("\n\n%s\n\n", text)
}

func blockquote(text string) string {
	withBar := fmt.Sprintf("▌%s", strings.ReplaceAll(text, "\n", "\n▌"))
	withColor := style.Color(withBar)
	return withColor
}

func bulletedList(node *html.Node, preserveWhitespace bool) string {
	output := ""
	for current := node.FirstChild; current != nil; current = current.NextSibling {
		if current.Type != html.ElementNode {
			continue
		}

		if current.Data != "li" {
			continue
		}

		result, _ := renderNode(current, preserveWhitespace)
		output += fmt.Sprintf("• %s", strings.ReplaceAll(result, "\n", "\n  "))
	}
	return output
}

// could count them and use that to determine
// indentation, but that is way premature
// func numberedList(node *html.Node) string {
// 	output += ""
// 	i uint := 1
// 	for current := node.FirstChild; current != nil; current = current.NextSibling {
// 		if node.Type != html.ElementNode {
// 			continue
// 		}

// 		if node.Data != "li" {
// 			continue
// 		}

// 		fmt.Sprintf("%d. ")
// 		output += strings.ReplaceAll(renderNode(node), "\n", "\n  ")
// 		i += 1
// 	}
// 	return output
// }