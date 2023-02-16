### `Object | Link`

For now I will not handle `Object | Link` types and will instead just filter down to `Post`/`Actor` and give warnings for `Link`s. I don't know how `Link`s would be displayed visually as, e.g., `inReplyTo` (although I see the utility). More importantly though, I don't know how to represent it (e.g. a `Link | Post` slice) in Go.

For the two `Image | Link` cases (`image` and `icon`), I will just convert `Images` to `Links` (using some algorithm to get the best URL, looking at mime types and resolutions). It maps over nicely. I'll do this in the `image`, `icon`, `attachment` scenarios, in addition to potentially others. Furthermore this logic can then be reused when I am pulling the URL from a `Video`, `Image`, etc.

Thinking now, all media types (`Document`, `Audio`, etc), will have a `Link()` function that returns the link using an algorithm for finding the best mime types and resolutions, based on whether it's a `Document`, `Audio`, etc. Then when I am, e.g., looking in `image`, `icon`, or `attachment`s, I will just loop through the list, keeping `Link`s and converting posts to `Link`s via `post.Link()` to keep the slice homogenous.

For the conversion it will be pulled from `name` of the parent document. So

`Document.Link()`, if the `Document.url` is a `Link` use it, if it's a string, then:
`Document.name` -> `Link.name`
`Document.url` -> `Link.href`
iff `content` and `summary` are absent: `Document.mediaType` -> `Link.mediaType` (Mastodon and PeerTube misuse it in this way, they use `mediaType` to refer to the `url`, not `content`)

## Bugs

By far the biggest flaw right now is that, if fulfilling an `id` results in an object with a re-retrieve condition (e.g. only having an `id` and `type`) the object will be re-retrieved infinitely. I need to add a flag "from source" (probably source as `nil`) to say to not re-retrieve.

## Improvements

Create a struct called `Fragment` that has `text string` for rendered text, `warnings []Warning` (or similar) for problems found during rendering, and `links map[rune]*url.URL` for hyperlinks that can be found within the text.

I need to redesign everything with warnings. And ban `nil` from, e.g. `string` types (if possible). Thus every return type is sane (e.g. empty string, empty struct). Only thing is time and URL must be pointers so they can be nilled. The only human rule is to never return `nil` in place of an interface.

## Future Plans

If the CLI client works out, look into making a really nice-looking plain-HTML/CSS front-end that it can serve.

## Ideas for related projects

* Static ActivityPub site generator based on Dhall
* A patch to Searx (or a better version) that serves results over ActivityPub to be read by an ActivityPub client
* A dedicated ActivityPub search engine to solve the supposed discoverability problem inherent to decentralized systems (yet doesn't exist with the web)
* Read-only RSS/Atom/JSON feed to ActivityPub hoster

## Misuse

Because ActivityPub supports `Article`s, `Note`s, `Image`s, having an account on PeerTube and PixelFed and Mastodon is pointless. I think people do this because of this mindset engrained over the past few years of wanting a big list of platforms that I am on. Other reason may be wanting to categorize your things, but clients should just be capable of filtering on `Note`s vs `Image`s to solve that problem.

Another misuse is organizations or people with websites using `mastodon.social`. The entire point of federation is to not use the centralized platform, but rather to use a host that makes sense for you.

## Quirks

For WebFinger, accept `application/json` in addition to `application/jrd+json`. Don't specify an `Accept` header because you don't need to. Problem found on PeerTube.

The thing where, if `content` is absent, `mediaType`/`name` apply to the `url`. Problem found on Mastodon, PeerTube, and PixelFed.

Future recommendation: add a `nameMediaType` field that applies to `name` so it can have markup. Default is `text/plain`. To add markup, recommendation is to use `text/markdown` so that it works fine on prior clients that treat it like `text/plain`.

Use both of the `Accept` headers, some sites only respond to `application/activity+json` (PixelFed).

## TODO

Document the reasoning for treating everything as JSON instead of JSON-LD.

## Minimal HTTPS

After learning HTTP it feels like HTTP1.0 and HTTP3 have good niches. HTTP1.0 is super simple, `Connection: close` by default, so it is delimited by TCP close (which is fine for JSON). It has no chunking, so that isn't a problem. One TCP connection per request. On the other hand, HTTP3 is a binary protocol. (Amazing to think that the entire Web has been run off of a text-based protocol. No wonder everything breaks all the time.) Also, HTTP3 itself seems relatively lightweight because it looks like lower-level stuff is in QUIC instead of jammed in the HTTP headers.

So I think ActivityPub clients should support HTTP1.0 for simple Gemini-style use-cases and hacking, and HTTP3 for more professional use-cases.
