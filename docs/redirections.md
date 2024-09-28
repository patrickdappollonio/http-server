# Redirections

- [Redirections](#redirections)
  - [Redirections and path prefix](#redirections-and-path-prefix)
  - [Syntax](#syntax)
    - [Exact match](#exact-match)
    - [Splat match](#splat-match)
    - [Path parameter match](#path-parameter-match)
    - [Querystring parameter match](#querystring-parameter-match)
    - [Maintaining querystring parameters](#maintaining-querystring-parameters)
    - [Escaping colons in URLs](#escaping-colons-in-urls)
  - [Inspecting redirections](#inspecting-redirections)

> [!WARNING]
> Redirections is a beta feature. Future versions of `http-server` may change the way redirections are handled. A given version of `http-server` will never change how redirections work, so if you want stability, consider pinning `http-server` to a specific version. Use it at your own risk.

`http-server` implements a basic redirection system that allows you to:

* Redirect requests to other locations, either permanently (HTTP status code `301`) or temporarily (HTTP status code `302`).
* Redirect using a rule system that allows:
  * Splat matching, which allows you to match any path after a certain point.
  * Exact matching, which allows you to match a specific path.
  * Parameter matching, which allows you to match a path with a specific parameter.
  * Querystring matching, which allows you to match a path with a specific querystring.
  * Querystring conversion matching, which allows you to match a querystring and redirect to a path without querystrings.

The rule system is inspired by the solutions currently available at [Cloudflare](https://developers.cloudflare.com/pages/configuration/redirects/) and [Netlify](https://docs.netlify.com/routing/redirects/) but they don't work necessarily the same way.

## Redirections and path prefix

> [!WARNING]
> The redirection system will ignore the value set on `--pathprefix` and will always redirect from the root of the server. This is intentional, to ensure that you can still serve content under a subpath, but you can still control the redirections from the root of the server.

Since `http-server` supports masking a path prefix for the folder you're in (for example, if you need to serve a folder in `example.com/blog`, you can start `http-server` with `--pathprefix=/blog` and the contents of the folder will be displayed only when accessing `example.com/blog`), **the redirections system does not take into account the path prefix**.

In other words, redirections work from the serving root of the server, not from the path prefix. This is to ensure that while you can still serve content under a subpath, you can still control the redirections from the root of the server.

By default, `http-server` when invoked with `--pathprefix=/blog` will still listen on the root of the server, but it will produce a redirection from `/` to `/blog` to ensure that the user is redirected to the correct path.

## Syntax

The syntax is quite simple, it follows the pattern:

```xml
[old] [new] [permanent|temporary]
```

Where:

* `[old]` is the path, relative to the root of `http-server` where the redirection should happen.
* `[new]` is the path where the request should be redirected to. This path can be relative to `http-server` or absolute to a different URL.
* `[permanent|temporary]` is the type of redirection. Use `permanent` for a `301` status code or `temporary` for a `302` status code.

Any value in the URL not covered by a match expression will be removed from the URL when redirecting. The same applies for querystring parameters.

### Exact match

The following example will redirect an exact match from `/old` to `/new`:

```bash
# redirects example.com/old to example.com/new
/old /new permanent
```

### Splat match

The following example will redirect any path that starts with `/old` to the same path in `/new`, so if you requested `/old/foo.txt`, you will land on `/new/foo.txt`:

```bash
# redirects example.com/old/* to example.com/new/*
/old/:splat /new/:splat permanent
```

If instead, the new location does not contain the exact same files as the old location, you can use the following syntax to redirect any prefix to a new location **without maintaining the same path**:

```bash
# redirects example.com/old/* to example.com/new
/old/* /new permanent
```

> [!TIP]
> You can use `:splat` when you want `http-server` to copy the path after the match to the new location, or `*` when you want to redirect to a new location without maintaining the same path.

### Path parameter match

To match a path parameter and redirect to a new location, you can name the given parameter like in many frameworks. The following example will redirect `/posts/:id` to `/articles/:id`:

```bash
# redirects example.com/posts/123 to example.com/articles/123
/posts/:id /articles/:id permanent
```

> [!WARNING]
> The parameter `:splat` is reserved for splat matching, so you can't use it as a parameter name.

### Querystring parameter match

There are two things you can do with querystring parameters: first, you can redirect a querystring parameter to a path parameter. The following example will redirect `/posts?id=123` to `/articles/123`:

```bash
# redirects example.com/posts?id=123 to example.com/articles/123
/posts?id=:id /articles/:id permanent
```

And second, you can redirect a querystring parameter to a new location with the same querystring parameter:

```bash
# redirects example.com/posts?id=123 to example.com/articles?id=123
/posts?id=:id /articles?id=:id permanent
```

### Maintaining querystring parameters

By default, the redirection system will remove any querystring parameter unless matched. You can prevent this behaviour with the special syntax `?!`, which will maintain non-conflictive querystring parameters.

To do so, append `?!` to the old path:

```bash
/posts/:id?! /articles/:id permanent
```

This will produce the following redirect:

```text
/posts/25?category=tech&utm_source=github --> /articles/25?category=tech&utm_source=github
```

There's one caveat: if the querystring parameter conflicts with the path parameter, **the unmatched querystring parameter will be removed**. Consider the following example:

```bash
/posts/:id?! /posts?id=:id temporary
```

This will redirect the following URL:

```
/posts/23?category=tech&id=9999 --> /posts?id=23&category=tech
```

Note how the duplicated parameter, `id=9999` was removed in favour of the parameter defined in the path. When maintaining querystring parameters with `?!`, only those parameters that don't conflict with the rules provided in the redirection matching will be maintained.

### Escaping colons in URLs

Since colons (`:`) play such an important part of the redirection engine, using them in a URL in a place that you might not want to match a parameter can be tricky. To avoid this, you can escape colons with a backslash (`\`), like so:

```bash
/tech\:articles/:id   /articles/tech/:id   permanent
```

This will redirect `/tech:articles/123` to `/articles/tech/123`.

## Inspecting redirections

`http-server` logs will report redirections. Consider the following redirections file:

```bash
/:splat https://www.patrickdap.com/:splat temporary
```

Making a `HTTP GET` request will produce the redirection:

```bash
$ curl -i http://localhost:1234/foo/bar/baz
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Etag: "427d467004d2337f70dac7618d9549b478dee0f3"
Location: https://www.patrickdap.com/foo/bar/baz
Date: Sat, 28 Sep 2024 02:33:32 GMT
Content-Length: 61

<a href="https://www.patrickdap.com/foo/bar/baz">Found</a>.
```

And the `http-server` logs will report the redirection:

```bash
2024/09/27 22:35:59 REDIR "/foo/bar/baz" -> "https://www.patrickdap.com/foo/bar/baz" (status: 302)
```
