# Servitor

A command line, RSS-style reader for the fediverse.

![](/assets/demo.gif)

# Features

* **Works with fediverse software such as Mastodon, Lemmy, PeerTube, and more**
  * microblogs like Linus Torvalds @torvalds@social.kernel.org
  * video channels like Luke Smith @luke@videos.lukesmith.xyz
  * subs like RPGMemes @rpgmemes@ttrpg.network
  * hashtags like tweesecake.social's [#linux](https://tweesecake.social/tags/linux)
* **Doesn't require a server**
* **Sidesteps defederation politics**
  * because posts are always pulled directly from their source site, you are not affected by servers blocking each other
* **Keeps your subscriptions private**
  * subscriptions are stored locally, so you can subscribe to people without sending them follow requests
* **Follows the Unix philosophy**
  * the browser itself is just handles the ActivityPub protocol, media viewing is all handled by separate programs

# Installation

TODO

# Usage

Use `servitor open @username@example.org` or `servitor open https://example.org/user/username` to open things.\
Use `servitor feed feed-name` to open feeds (see below).

## Configuration

The config file is located at `~/.config/servitor/config.toml`.

```toml
[feeds]
# the entries will be spliced together in reverse-chronological order, just like an RSS reader
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
preload_amount = 5 # the amount of posts to load in above and below the highlighted post
timeout_seconds = 5
cache_size = 128 # the number of JSON responses the cache can hold

[media]
# described below
```

### Media Hook

There are various ways to open files on Linux (`xdg-open`, `mailcap`, [`handlr`](https://github.com/chmln/handlr), bespoke scripts, etc). The `media.hook` config option allows you to configure whichever one you use. The value is a list of strings that will be executed as a command. Parameters will be substituted as follows:

* `%url` &mdash; substituted with the URL being opened
* `%mimetype` &mdash; substituted with the media type being opened, e.g. `image/png`
* `%supertype` &mdash; substituted with the first part of the media type, e.g. `image`
* `%subtype` &mdash; substituted with the subtype, e.g. `png`

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

# Where to Find Content to Follow

* [Unofficial Subreddit Migration List (Lemmy, Kbin)](https://www.quippd.com/writing/2023/06/15/unofficial-subreddit-migration-list-lemmy-kbin-etc.html)
* [fedi.directory](https://fedi.directory/)

Please submit a PR if you know of another good resource.

# A Brief Overview of the Fediverse

For the purpose of this browser, the fediverse can be thought of as a collection of internet forums that use a shared protocol called ActivityPub. Instead of serving content over `text/html`, they serve their content over `application/activity+json`, which provides for higher-level semantics such as comment sections, retweets, etc. Just like conventional forums, each site has a different moderation policy, and the administrators of each site have complete control over the moderation of that site. Unlike conventional internet forums, the fediverse allows users from one site to interact with (like, follow, comment on, etc) users and posts on another site, assuming the administrators of the both sites permit the interaction.

# Supported Markup Formats

Servitor can render posts published in:
* [HTML](https://en.wikipedia.org/wiki/HTML)
* [gemtext](https://gemini.circumlunar.space/docs/gemtext.gmi)
* plain text
* [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax)

# Dependencies

I tried to minimize the amount of dependencies used. They are:

* [BurntSushi/toml](https://github.com/BurntSushi/toml) for parsing the config file
* [yuin/goldmark](https://github.com/yuin/goldmark) for rendering posts published in Markdown (currently the only software I'm aware of that serves posts with Markdown is PeerTube)
* [hashicorp/golang-lru/v2](https://github.com/hashicorp/golang-lru) for the local cache

# Planned Improvements

* The ability to connect to ActivityPub search engines and algorithm providers (however controversial that may be)
* Better support for long-form content
