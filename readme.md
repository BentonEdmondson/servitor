# Servitor

A command-line Fediverse client that doesn’t require a server.

![](/assets/demo.gif)

# Features

* **Works with fediverse software such as Mastodon, Lemmy, PeerTube, and more**
  * microblogs like Linus Torvalds [@torvalds@social.kernel.org](https://social.kernel.org/torvalds)
  * video channels like Luke Smith [@luke@videos.lukesmith.xyz](https://videos.lukesmith.xyz/a/luke/video-channels)
  * subs like RPGMemes [@rpgmemes@ttrpg.network](https://ttrpg.network/c/rpgmemes)
  * hashtags like tweesecake.social's [#linux](https://tweesecake.social/tags/linux)
* **Doesn't require a server**
* **Sidesteps defederation politics**
  * because posts are always pulled directly from their source site, you are not affected by servers blocking each other
* **Keeps your subscriptions private**
  * subscriptions are stored locally, so you can subscribe to people without sending them follow requests
* **Follows the Unix philosophy**
  * the browser itself is just handles the ActivityPub protocol, media viewing is all handled by separate programs

# Installation

1. Run `uname -ms` and, based on the output, download the latest corresponding [release](https://github.com/BentonEdmondson/servitor/releases).
1. `cd` into the directory that servitor is in (e.g. `cd ~/Downloads`).
1. Run `mv servitor-version-arch-os servitor` to rename the binary to `servitor`.
1. Run `chmod +x servitor` to make it executable.
1. Run `./servitor open @Gargron@mastodon.social`, for example, to try it out.
1. Run `mv servitor ~/.local/bin` to allow it to be run from anywhere (might not work on some distributions).

I only test `Linux x86_64` releases.

# Usage

* `servitor open @username@example.org` to open profiles.
* `servitor open https://example.org/user/username` to open links.
* `servitor feed feed-name` to open feeds (see below).

## Configuration

The config file is located at `~/.config/servitor/config.toml`.

```toml
[feeds]
# the entries will be spliced together
# in chronological order, like an RSS reader
linux = [ # open with `servitor feed linux`
    "@torvalds@social.kernel.org",
    "@luke@videos.lukesmith.xyz",
    "@thelinuxexperiment@tilvids.com",
]
dnd = [ # open with `servitor feed dnd`
    "@rpgmemes@ttrpg.network",
    "@dnd@lemmy.world",
]

[style.colors]
primary = "#A4f59b"
error = "#9c3535"
highlight = "#0d7d00"
code_background = "#4b4b4b"

[network]
preload_amount = 5 # the number of posts to load in above and below the highlighted post
timeout_seconds = 5
cache_size = 128 # the number of JSON responses the cache can hold

[media]
# described below
```

### Media Hook

There are various ways to open files on Linux (`xdg-open`, `mailcap`, [`handlr`](https://github.com/chmln/handlr), bespoke scripts, etc). The `media.hook` config option allows you to configure whichever one you use. The value is a list of strings that will be executed as a command. Parameters will be substituted as follows:

* `%url` — substituted with the URL being opened
* `%mimetype` — substituted with the media type being opened, e.g. `image/png`
* `%supertype` — substituted with the first part of the media type, e.g. `image`
* `%subtype` — substituted with the subtype, e.g. `png`

Here is a simple example config that opens videos and gifs in `mpv`, images in `feh`, and everything else in `firefox`:

```toml
[media]
hook = [
    "sh", "-c",
    '''
        if test "$2" = "video"; then exec mpv --keep-open=yes "$0"; fi
        if test "$1" = "image/gif"; then exec mpv --keep-open=yes "$0"; fi
        if test "$2" = "image"; then exec feh --scale-down --image-bg black "$0"; fi
        exec firefox "$0"
    ''',
    "%url", "%mimetype", "%supertype"
]
```

`media.hook` defaults to `["xdg-open", "%url"]`.

## Keybindings

### Navigation
`j` — move down\
`k` — move up\
`space` — select the highlighted item\
`c` — view the creator of the highlighted item\
`r` — view the recipient of the highlighted item (e.g. the group it was posted to)\
`a` — view the actor of the activity (e.g. view the retweeter of a retweet)\
`h` — move back in your browser history\
`l` — move forward in your browser history\
`g` — move to the expanded item (i.e. move to the current OP)\
`ctrl+c` — exit the program

### Media
`p` — open the highlighted user's profile picture\
`b` — open the highlighted user's banner\
`o` — open the content of a post itself (e.g. open the video associated with a video post)\
number keys — open a link within the highlighted text

# Contributing

## Building

```
CGO_ENABLED=0 go build -ldflags "-X main.version=$(git rev-parse --short HEAD)"
```

## Updating

```
go get -u all && go mod tidy
```

## Testing

```
go test ./...
```

## Formatting

```
go fmt ./...
```

# Where to Find Content to Follow

* [Unofficial Subreddit Migration List (Lemmy, Kbin)](https://www.quippd.com/writing/2023/06/15/unofficial-subreddit-migration-list-lemmy-kbin-etc.html)
* [fedi.directory](https://fedi.directory/)

Please submit a PR if you know of another good resource.

# A Brief Overview of the Fediverse

For the purpose of this browser, the fediverse can be thought of as a collection of internet forums that use a shared protocol called ActivityPub. Instead of serving content over `text/html`, they serve their content over `application/activity+json`, which provides for higher-level semantics such as comment sections, retweets, etc.

Just like conventional forums, each site has a different moderation policy, and the administrators of each site have complete control over the moderation of that site. Unlike conventional internet forums, the fediverse allows users from one site to interact with (like, follow, comment on, etc) users and posts on another site, assuming the administrators of the both sites permit the interaction.

# Supported Markup Formats

Servitor can render posts published in:
* [HTML](https://en.wikipedia.org/wiki/HTML)
* [gemtext](https://gemini.circumlunar.space/docs/gemtext.gmi)
* plain text
* [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax)

# Dependencies

I tried to minimize the number of dependencies used. They are:

* [BurntSushi/toml](https://github.com/BurntSushi/toml) for parsing the config file
* [yuin/goldmark](https://github.com/yuin/goldmark) for rendering posts published in Markdown (currently the only software I'm aware of that serves posts with Markdown is PeerTube)
* [hashicorp/golang-lru/v2](https://github.com/hashicorp/golang-lru) for the local cache
