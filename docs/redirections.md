# Redirections

- [Redirections](#redirections)
  - [Redirections and path prefix](#redirections-and-path-prefix)
  - [Syntax](#syntax)
    - [Exact match](#exact-match)
    - [Splat match](#splat-match)
    - [Path parameter match](#path-parameter-match)
    - [Querystring parameter match](#querystring-parameter-match)
    - [Maintaining querystring parameters](#maintaining-querystring-parameters)
    - [Regular expression match](#regular-expression-match)
      - [Examples](#examples)
      - [Validation of placeholders](#validation-of-placeholders)
      - [Limitations](#limitations)
    - [Escaping special characters](#escaping-special-characters)
  - [Inspecting redirections](#inspecting-redirections)

> [!WARNING]
> Redirections is a beta feature. Future versions of `http-server` may change the way redirections are handled. A given version of `http-server` will never change how redirections work, so if you want stability, consider pinning `http-server` to a specific version. Use it at your own risk.

`http-server` implements a basic redirection system that allows you to:

* Redirect requests to other locations, either permanently (HTTP status code `301`) or temporarily (HTTP status code `302`).
* Redirect using a rule system that allows:
  * **Exact matching**, which allows you to match a specific path.
  * **Splat matching**, which allows you to match any path after a certain point.
  * **Parameter matching**, which allows you to match a path with specific parameters.
  * **Querystring matching**, which allows you to match a path with specific query parameters.
  * **Regular expression matching**, which allows you to define complex matching patterns.
  * **Querystring conversion matching**, which allows you to match a querystring and redirect to a path without querystrings.

The rule system is inspired by solutions currently available at [Cloudflare](https://developers.cloudflare.com/pages/configuration/redirects/) and [Netlify](https://docs.netlify.com/routing/redirects/), but they don't necessarily work the same way.

Redirections are evaluated on startup, and any errors in the redirections file will prevent `http-server` from starting. This ensures that you can't accidentally introduce a broken redirection rule. This also means that **you can't change redirections without restarting `http-server`**.

## Redirections and path prefix

> [!WARNING]
> The redirection system will ignore the value set on `--pathprefix` and will always redirect from the root of the server. This is intentional, to ensure that you can still serve content under a subpath, but control the redirections from the root of the server.

Since `http-server` supports masking a path prefix for the folder you're in (for example, if you need to serve a folder at `example.com/blog`, you can start `http-server` with `--pathprefix=/blog`, and the contents of the folder will be displayed only when accessing `example.com/blog`), **the redirection system does not take into account the path prefix**.

In other words, redirections work from the serving root of the server, not from the path prefix. This ensures that while you can serve content under a subpath, you can still control the redirections from the root of the server.

By default, when invoked with `--pathprefix=/blog`, `http-server` will still listen on the root of the server but will produce a redirection from `/` to `/blog` to ensure that the user is redirected to the correct path.

## Syntax

The syntax is quite simple and follows the pattern:

```bash
[old] [new] [permanent|temporary]
```

Where:

* `[old]` is the path, relative to the root of `http-server`, where the redirection should happen.
* `[new]` is the path where the request should be redirected to. This path can be relative to `http-server` or an absolute URL.
* `[permanent|temporary]` is the type of redirection. Use `permanent` for a `301` status code or `temporary` for a `302` status code.

Any value in the URL not covered by a match expression will be removed from the URL when redirecting. The same applies to querystring parameters.

### Exact match

The following example will redirect an exact match from `/old` to `/new`:

```bash
# Redirects example.com/old to example.com/new
/old /new permanent
```

### Splat match

The following example will redirect any path that starts with `/old` to the same path in `/new`. For instance, if you request `/old/foo.txt`, you will be redirected to `/new/foo.txt`:

```bash
# Redirects example.com/old/* to example.com/new/*
/old/:splat /new/:splat permanent
```

If the new location does not contain the same path structure as the old location, you can redirect any prefix to a new location **without maintaining the same path**:

```bash
# Redirects example.com/old/* to example.com/new
/old/* /new permanent
```

> [!TIP]
> Use `:splat` when you want `http-server` to copy the path after the match to the new location, or `*` when you want to redirect to a new location without maintaining the same path.

### Path parameter match

To match a path parameter and redirect to a new location, you can name the given parameter as in many web frameworks. The following example will redirect `/posts/:id` to `/articles/:id`:

```bash
# Redirects example.com/posts/123 to example.com/articles/123
/posts/:id /articles/:id permanent
```

> [!WARNING]
> The parameter `:splat` is reserved for splat matching, so you can't use it as a parameter name.

### Querystring parameter match

You can redirect a querystring parameter to a path parameter. The following example will redirect `/posts?id=123` to `/articles/123`:

```bash
# Redirects example.com/posts?id=123 to example.com/articles/123
/posts?id=:id /articles/:id permanent
```

Alternatively, you can redirect a querystring parameter to a new location while maintaining the same querystring parameter:

```bash
# Redirects example.com/posts?id=123 to example.com/articles?id=123
/posts?id=:id /articles?id=:id permanent
```

### Maintaining querystring parameters

By default, the redirection system will remove any querystring parameter unless matched. You can prevent this behavior with the special syntax `?!`, which will maintain non-conflicting querystring parameters.

To do so, append `?!` to the old path:

```bash
/posts/:id?! /articles/:id permanent
```

This will produce the following redirect:

```text
/posts/25?category=tech&utm_source=github --> /articles/25?category=tech&utm_source=github
```

**Note:** If a querystring parameter conflicts with a path parameter, **the unmatched querystring parameter will be removed**. For example:

```bash
/posts/:id?! /posts?id=:id temporary
```

This will redirect:

```
/posts/23?category=tech&id=9999 --> /posts?id=23&category=tech
```

Notice how the duplicated parameter `id=9999` was removed in favor of the parameter defined in the path. When maintaining querystring parameters with `?!`, only those parameters that don't conflict with the rules provided in the redirection matching will be maintained.

### Regular expression match

In addition to the standard redirection rules, `http-server` supports **regex-based redirection rules**. This allows you to define more complex matching patterns using regular expressions.

The syntax for regex-based rules is:

```bash
regexp "<pattern>" "<replacement>" [permanent|temporary]
```

Where:

- `regexp` indicates that this is a regex-based rule.
- `<pattern>` is a regular expression pattern, enclosed in double quotes (`"`). This pattern is applied to the entire request URI, which includes the path and the query string.
- `<replacement>` is the replacement string, enclosed in double quotes (`"`). It can include references to captured groups from the pattern.
- `[permanent|temporary]` is the type of redirection. Use `permanent` for a `301` status code or `temporary` for a `302` status code.

**Important Notes:**

- **Enclose patterns and replacements in double quotes.** If you need to include a double quote within the pattern or replacement, escape it with a backslash (`\"`).
- **Capture groups** in the pattern can be referenced in the replacement using:
  - `$1`, `$2`, etc., for **positional groups**.
  - `$name` for **named capture groups** defined with `(?P<name>...)`.
- **Regex-based rules are self-contained.** They do not mix with the placeholder-based logic (`:param` or `*`) or the `?!` syntax for maintaining query parameters.
- **Query parameters are handled entirely within the regex pattern and replacement.** If you need to match or include query parameters, include them in your regex pattern and replacement.

#### Examples

**Redirect with a positional capture group**

The following example redirects any path starting with `/blog/` to `/articles/`, preserving the rest of the path:

```bash
# Redirects example.com/blog/* to example.com/articles/*
regexp "^/blog/(.+)$" "/articles/$1" permanent
```

- Requesting `/blog/my-first-post` will redirect to `/articles/my-first-post`.

**Redirect with a named capture group**

You can use named capture groups in your regex pattern and reference them in the replacement:

```bash
# Redirects example.com/user/:username to example.com/profile/:username
regexp "^/user/(?P<username>[^/]+)$" "/profile/$username" temporary
```

- Requesting `/user/johndoe` will redirect to `/profile/johndoe`.

**Including query parameters**

If you need to match or include query parameters, include them in your regex pattern and replacement:

```bash
# Redirects example.com/search?q=term to example.com/find?q=term
regexp "^/search\\?q=(.+)$" "/find?q=$1" temporary
```

- Requesting `/search?q=golang` will redirect to `/find?q=golang`.

**Handling complex patterns**

You can define more complex patterns using regular expressions:

```bash
# Redirects example.com/order/:orderId/item/:itemId to example.com/orders/:orderId/items/:itemId
regexp "^/order/(?P<orderId>\\d+)/item/(?P<itemId>\\d+)$" "/orders/$orderId/items/$itemId" permanent
```

- Requesting `/order/123/item/456` will redirect to `/orders/123/items/456`.

#### Validation of placeholders

When using placeholders in the replacement string (e.g., `$1`, `$username`), ensure that they correspond to actual capture groups defined in your regex pattern. If a placeholder does not match any capture group, the redirection rule will be invalid, and `http-server` will report an error during startup.

**Invalid example with unmatched named capture group:**

```bash
# Invalid rule: $username is not defined in the pattern
regexp "^/user/(?P<userid>[^/]+)$" "/profile/$username" temporary
```

- This will result in an error: `undefined placeholder "$username" in replacement on line X`.

**Invalid example with unmatched positional capture group:**

```bash
# Invalid rule: There is no second positional group ($2)
regexp "^/user/(.+)$" "/profile/$2" temporary
```

- This will result in an error: `undefined placeholder "$2" in replacement on line X`.

#### Limitations

- The `?!` syntax for maintaining query parameters **does not apply** to regex-based rules.
- Regex-based rules are evaluated in the order they appear in your redirections file, just like non-regex rules.
- Regex patterns are applied to the **entire request URI**, including the path and query string.

### Escaping special characters

**In non-regex rules:**

Since colons (`:`) play an important part in the redirection engine for defining parameters, using them in a URL where you don't want to match a parameter can be tricky. To avoid this, you can escape colons with a backslash (`\`):

```bash
# Redirects example.com/tech:articles/123 to example.com/articles/tech/123
/tech\:articles/:id /articles/tech/:id permanent
```

**In regex-based rules:**

In regex patterns, you need to escape special regex characters according to regular expression syntax. For example, to match a literal question mark (`?`), you need to escape it with a backslash (`\\?`):

```bash
# Redirects example.com/search?q=term to example.com/find?q=term
regexp "^/search\\?q=(.+)$" "/find?q=$1" temporary
```

If you need to include a backslash or a double quote in your pattern or replacement, escape it with another backslash:

- To include a double quote (`"`), use `\"`.
- To include a backslash (`\`), use `\\`.

**Example with escaped double quotes:**

```bash
# Redirects example.com/say/"hello world" to example.com/quote/hello world
regexp "^/say/\"(.+)\"$" "/quote/$1" temporary
```

## Inspecting redirections

`http-server` logs will report redirections. Consider the following redirections file:

```bash
/:splat https://www.example.com/:splat temporary
```

Making an HTTP GET request will produce the redirection:

```bash
$ curl -i http://localhost:1234/foo/bar/baz
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Etag: "427d467004d2337f70dac7618d9549b478dee0f3"
Location: https://www.example.com/foo/bar/baz
Date: Sat, 28 Sep 2024 02:33:32 GMT
Content-Length: 61

<a href="https://www.example.com/foo/bar/baz">Found</a>.
```

And the `http-server` logs will report the redirection:

```bash
2024/09/27 22:35:59 REDIR "/foo/bar/baz" -> "https://www.example.com/foo/bar/baz" (status: 302)
```
