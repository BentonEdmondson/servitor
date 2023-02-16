# Presentational Subset of HTML

It is popular to suggest that HTML markup should be semantic and presentational styling should be applied near-exclusively by stylesheets. In practice, this is not how HTML is used. For instance, people nearly never refer to `*` or `_` as "strong importance, seriousness, or urgency" or "stress emphasis" markers. They refer to them as bold and italics and use them that way. When people write an article on Medium or Google Docs and click the `bold` button, they don't think "oh this text is of strong seriousness and should be marked as such"; they want it to be a thicker, perhaps darker, font, i.e. bold.

## Non-Textual Elements

* `img` for images.
* `video` for videos (content between the tags is ignored).
* `audio` for audio (content between the tags is ignored).
* `iframe` for including other content like YouTube videos, Tweets, etc.
* `hr` for a [section break](https://en.wikipedia.org/wiki/Section_(typography)#Flourished_section_breaks).

## Inline Elements

* `a` for links.
* `s` for strikethrough. `del` as an alias for compatibility.
* `code` for monospace font. Like backticks in Markdown.
* `i` for italics. `em` as an alias for compatibility.
* `b` for bold. `strong` as an alias for compatibility.
* `u` for underline.
* `mark` for highlight. It should be rendered with a changed background color, but the color itself is not specified.
* `span` is ignored for compatibility.
* `br` for a single line break.

## Block Elements

All block elements are separated from adjacent content with a blank line.

* `p` for a paragraph. Visually it does nothing besides make the contents a block (and thus separated from adjacent content). `div` as an alias for compatibility.
* `pre` for text rendered in a monospace font. Like a code block (three backticks) in Markdown. Placing the `code` element within `pre` does nothing as `pre` already monospaces the content.
* `blockquote` for an indented and recolored quote.
* `ol` for a numbered list.
* `ul` for a bullet list.
    * `li` for the elements enumerated within `ol` or `ul`. If not the immediate child of `ol` or `ul`, then it is treated like `span` and ignored.
* `h1`, `h2`, `h3`, `h4`, `h5`, `h6` for headers as used in Markdown.

## Styles

This involves the following styles:
* Code - monospace with grey background
* Highlight - brighter background
* Strikethrough
* Underline
* Bold
* Italics

## Formatters

* Block
* Header
* Bullet
* Number

## Whitespace

Currently the plan with whitespace is to
* Leave code and pre content alone
* Collapse all other newline-containing whitespace into a single newline
* Collapse all non-newline-containing whitespace into a single space
* Prepend and append two newlines to each block
* Collapse all sequences of over two newlines into two newlines
* Trim all newlines from the beginning and end of the document.

---

Reasons Semantic is dead:
* What are we going to mark sentences as sentences, questions as questions, etc?
* People clearly use markup for visuals, not semantic.