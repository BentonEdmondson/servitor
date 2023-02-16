# Implementation Guide

ActivityPub's spec isn't specific in certain areas, so this document describes common usage.

## Rendering Markup

`text/plain` is easy to render. Potentially you can go through and try to identify links to make them easily selectable.

`text/gemini` is not used by anyone but easy to render and nice for Twitter-style content.

`text/html`: I will support:

From section 4.8: img, audio, video, iframe for image/*, audio/\*, video/\*, and text/html.
From section 4.6: a
From section 4.5: most of the elements
From section 4.4: most of the elements
From section 4.3: h1, h2, etc.