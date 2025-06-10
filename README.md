# `http-server`, a simple HTTP file server

[![Github Downloads](https://img.shields.io/github/downloads/patrickdappollonio/http-server/total?color=orange&label=github%20downloads)](https://github.com/patrickdappollonio/http-server/releases)
[![Container Downloads](https://img.shields.io/badge/container%20downloads-250k-orange)](https://github.com/users/patrickdappollonio/packages/container/package/docker-http-server) [![Go Report Card](https://goreportcard.com/badge/github.com/patrickdappollonio/http-server)](https://goreportcard.com/report/github.com/patrickdappollonio/http-server)

<img src="internal/server/assets/file-server.svg" width="160" align="right" /> `http-server` is a static file server with zero dependencies: **just one binary to run**. It also supports:

* **CORS support:** by setting the `Access-Control-Allow-Origin` header to `*`. `HEAD` requests, although unnecessary when doing CORS on `GET` requests, are also supported.
* **Authentication support:** via either plain username and password or through a JWT token, with optional support for validating if the token isn't expired.
* **Directory listing:** if no `index.html` or `index.htm` files are present in the directory, a directory listing page will show instead.
* **Markdown support:** if a `README.md` or `readme.md` file is present in the directory during directory listing, it will be rendered as HTML. Additional support for GitHub-flavored markdown is also available.
* **Fully air-gapped:** the directory listing feature is fully air-gapped, meaning that it does not require any external resources to be loaded. This is useful for environments where internet access is not available.
* **Redirections support:** if a `_redirections` file exists in the target directory, it will be used to redirect requests to other locations. Learn about the syntax [in the docs](docs/redirections.md).

The app is available both as a standalone binary and as a Docker container image.

For an example site built with `http-server`, check the demo page: [`https://patrickdappollonio.github.io/http-server-testing-page/`](https://patrickdappollonio.github.io/http-server-testing-page/). The [repository for the demo page](https://github.com/patrickdappollonio/http-server-testing-page) contains the folder being served.

### Docker image

[Find the latest version available here](https://github.com/users/patrickdappollonio/packages/container/package/docker-http-server). `latest` will always map to the latest version, which could led you to download a newer major version that might contain a breaking change. I recommend using `v2` for the tag, since it will always map to a stable version with all potential patches applied. This is the safest way to use to avoid any breaking changes.

```bash
# stable version 2
docker pull ghcr.io/patrickdappollonio/docker-http-server:v2

# pin to specific version
docker pull ghcr.io/patrickdappollonio/docker-http-server:v2.0.0

# bleeding edge version (will always update to latest)
docker pull ghcr.io/patrickdappollonio/docker-http-server:latest
```

Usage with `docker compose`:

```yaml
services:
  http-server:
    image: ghcr.io/patrickdappollonio/docker-http-server:v2
    restart: unless-stopped
    ports:
      - "5000:5000"
    volumes:
      - ./:/html:ro # Use the current directory as the source
    environment:
      - PORT=5000 # Configures the port (the default is 5000)
      - TITLE=HTTP Server Test Site # Changes the site name
      - CORS=true # Enable CORS for testing purposes
      # Other configuration options:
      # - BANNER=Welcome to the HTTP Server Test Site! # Changes the banner
```

#### Configuring the container

There are three ways to configure the container:

* **Using a YAML configuration file:** create a configuration file named `.http-server.yaml`. This file cannot be accessed using the file explorer mode nor it will show up in the directory listing. The variable names match the command line flags. For example, to set `--disable-markdown`, you can use `disable-markdown: true` in the configuration file.
* **Using environment variables:** The environment variables match the command line flags. For example, to set `--disable-markdown`, you can use `DISABLE_MARKDOWN=true` as an environment variable. Additionally, and to avoid collisions, all environment variables can be prefixed with `FILE_SERVER_`. For example, to set `--path` parameter, which would collide with your Operating System's `$PATH`, you can use instead `FILE_SERVER_PATH`.
* **Overwriting the `command` and `args`:** Overriding the arguments passed to the container is also possible. For Docker, see [overriding `CMD`](https://docs.docker.com/engine/reference/run/#cmd-default-command-or-options) but keep the `ENTRYPOINT` intact. For `docker compose`, see [overriding the `command`](https://docs.docker.com/compose/compose-file/#command). For Kubernetes, see [overriding `command` and `args`](https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/).

### Static binary

You can download the latest version from the [Releases page](https://github.com/patrickdappollonio/http-server/releases).

### Homebrew

You can also install the app using Homebrew. A Homebrew tap is available with the latest version:

```bash
brew install patrickdappollonio/tap/http-server
```

#### Usage options

The following options are available.

```text
A simple HTTP server and a directory listing tool.

Usage:
  http-server [flags]

Flags:
      --banner string                       markdown text to be rendered at the top of the directory listing page
      --cors                                enable CORS support by setting the "Access-Control-Allow-Origin" header to "*"
      --custom-404 string                   custom "page not found" to serve
      --custom-404-code int                 custom status code for pages not found
      --custom-css-file string              path within the served files to a custom CSS file
      --disable-cache-buster                disable the cache buster for assets from the directory listing feature
      --disable-directory-listing           disable the directory listing feature and return 404s for directories without index
      --disable-etag                        disable etag header generation
      --disable-markdown                    disable the markdown rendering feature
      --disable-redirects                   disable redirection file handling
      --ensure-unexpired-jwt                enable time validation for JWT claims "exp" and "nbf"
      --etag-max-size string                maximum size for etag header generation, where bigger size = more memory usage (default "5M")
      --force-download-extensions strings   file extensions that should be downloaded instead of displayed in browser
      --gzip                                enable gzip compression for supported content-types
  -h, --help                                help for http-server
      --hide-files-in-markdown              hide file and directory listing in markdown rendering
      --hide-links                          hide the links to this project's source code visible in the header and footer
      --jwt-key string                      signing key for JWT authentication
      --markdown-before-dir                 render markdown content before the directory listing
      --password string                     password for basic authentication
  -d, --path string                         path to the directory you want to serve (default "./")
      --pathprefix string                   path prefix for the URL where the server will listen on (default "/")
  -p, --port int                            port to configure the server to listen on (default 5000)
      --render-all-markdown                 if enabled, all Markdown files will be rendered using the same rendering as the directory listing READMEs
      --title string                        title of the directory listing page
      --username string                     username for basic authentication
  -v, --version                             version for http-server
```

### Detailed configuration

All the available configuration options are documented in the docs. You can find them [here](docs/).
