# JTP: JSON Transfer Protocol

JTP is a profile (in the same sense as RFC 3339) of HTTP/1.0 that facilitates extremely simple transfer of JSON objects or arrays over a TCP connection.

## Motivation

I had read about the `gemini://` protocol as a simplified alternative to `https://`, and so wanted to explore how complicated `https://` really was. Ultimately, I found that the subset of features needed to do a basic JSON transfer is extremely simple.

## Request

The format of the request I send is

```
GET /resource.json HTTP/1.0
Host: server.org
Accept: application/activity+json

```

I specify the version as `HTTP/1.0` to prevent the server from sending me a chunked-transfer encoded response. I specify `Host:` because, although not formally required in `HTTP/1.0`, many servers complain about its absence. `Accept:` is how I request that the server send me ActivityPub JSON instead of `text/html`.

## Responses

My response parsing handles any valid response to an `HTTP/1.0` request, but the most minimal responses I accept are as follows.

In the successful case:

```
HTTP/1.1 200
content-type: application/activity+json

{
    "the": "json"
}
```

In the redirect case:

```
HTTP/1.1 300
location: /over-here.json

```

And in the error case:

```
HTTP/1.1 400

```
