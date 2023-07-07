# Servitor

A command line, RSS-style reader for the fediverse.

// image

* **Works with fediverse software such as Mastodon, Lemmy, PeerTube, and more**: This browser works with all fediverse software that implements ActivityPub. You can view and subscribe to microblogs (like Linus Torvalds @torvalds@social.kernel.org), video channels (like Luke Smith @luke@videos.lukesmith.xyz), subs (like RPGMemes @rpgmemes@ttrpg.network), and hashtags (like tweesecake.social's [#linux](https://tweesecake.social/tags/linux)) all in the same reader.
* **Doesn't require a server**: This browser does not rely on you having a server, so you don't have to host your own server or find another one to rely on.
* **Sidesteps defederation politics**: Posts are always pulled directly from their source site, so you are not affected by servers blocking each other.
* **Keeps your subscriptions private**: Just like an RSS reader, subscriptions are stored locally, so you can subscribe to people without sending them follow requests.

# Usage

## Configuration

The config file is located at `~/.config/servitor/config.toml`.

```toml
[feeds]
# each entry is list of profiles to subscribe to
linux = [ # open with `servitor feed linux`
    "@torvalds@social.kernel.org",
    "@luke@videos.lukesmith.xyz",
    "@thelinuxexperiment@tilvids.com",
]
dnd = [ # open with `servitor feed dnd`
    "@rpgmemes@ttrpg.network",
    "@dnd@lemmy.world",
]

[media]
# the command that is called to open external media
# %u is automatically substituted with the url, %m is substituted with the mime type
hook = [ "xdg-open", "%u" ]
```

# Where to Find Content to Follow

* [Unofficial Subreddit Migration List (Lemmy, Kbin)](https://www.quippd.com/writing/2023/06/15/unofficial-subreddit-migration-list-lemmy-kbin-etc.html)
* [fedi.directory](https://fedi.directory/)

Please submit a PR if you know of another good resource.

# A Brief Overview of ActivityPub

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
