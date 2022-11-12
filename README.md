# Simple HTTP file server

[![Github Downloads](https://img.shields.io/github/downloads/patrickdappollonio/http-server/total?color=orange&label=github%20downloads)](https://github.com/patrickdappollonio/http-server/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/patrickdappollonio/http-server)](https://goreportcard.com/report/github.com/patrickdappollonio/http-server)

- [Simple HTTP file server](#simple-http-file-server)
  - [Usage](#usage)
  - [Features](#features)
    - [Static file server](#static-file-server)
    - [Directory listing](#directory-listing)
      - [File highlighting](#file-highlighting)
      - [Markdown rendering](#markdown-rendering)
      - [Air-gapped environment](#air-gapped-environment)
    - [Authentication](#authentication)
      - [Basic authentication](#basic-authentication)
      - [JWT authentication](#jwt-authentication)
    - [Configuration provided via config files](#configuration-provided-via-config-files)

`http-server` is a simple binary to provide a static HTTP server from a given folder listening by default on port `5000`.

There are multiple configuration options available, and you can see them by running `http-server --help`:

```text
A simple HTTP server and a directory listing tool.

Usage:
  http-server [flags]

Flags:
      --cors                   enable CORS support by setting the "Access-Control-Allow-Origin" header to "*"
      --disable-cache-buster   disable the cache buster for assets from the directory listing feature
      --disable-markdown       disable the markdown rendering feature
      --ensure-unexpired-jwt   enable time validation for JWT claims "exp" and "nbf"
  -h, --help                   help for http-server
      --hide-links             hide the links to this project's source code
      --jwt-key string         signing key for JWT authentication
      --markdown-before-dir    render markdown content before the directory listing
      --password string        password for basic authentication
  -d, --path string            path to the directory you want to serve (default "./")
      --pathprefix string      path prefix for the URL where the server will listen on (default "/")
  -p, --port int               port to configure the server to listen on (default 5000)
      --title string           title of the directory listing page
      --username string        username for basic authentication
  -v, --version                version for http-server
```

## Usage

There are two ways to use `http-server`: by either [downloading a static binary from the "Releases" page](https://github.com/patrickdappollonio/http-server/releases), or by using the [Docker container image](https://github.com/users/patrickdappollonio/packages/container/package/docker-http-server).

The container image can be pulled from:

```bash
docker pull ghcr.io/patrickdappollonio/docker-http-server:latest
```

You can use the container in a custom image by providing a `WORKDIR`, since the current working directory is served by default:

```dockerfile
FROM ghcr.io/patrickdappollonio/docker-http-server:latest
WORKDIR /html
```

Alternatively, you can use the container directly and point the path to the `/html` directory or any other directory:

```bash
docker run -p 5000:5000 -v /path/to/files:/html ghcr.io/patrickdappollonio/docker-http-server:latest --path /html
```

## Features

### Static file server

By default, the `http-server` will serve the files from the current directory and will listen on port `5000`. You can change the port by using the `--port` flag, and the path by using the `--path` flag.

### Directory listing

If the directory does not contain any `index.html` or `index.htm` file, the `http-server` will render a directory listing page, which will list all the files and folders in the directory, excluding `http-server`'s own configuration files.

Old versions of this program would render the contents in the `/html` folder, however, this was changed to a more dynamic approach, since `/html` was not a cross-platform suitable configuration value.

To allow for directory listing mode, folders cannot contain a folder called `_` (underscore). `http-server` reserves underscore folders for the directory listing mode and its assets.

A sample preview of the directory listing page can be seen [in this screenshot](img/sample-site.png) or in the live site [here](#).

#### File highlighting

`http-server` includes support for highlighting files in the UI. Useful when you want to direct the end user's attention toward one or more files being shown via the directory listing feature. To "highlight" a file, `CTRL+Shift+Click` (or `CMD+Shift+Click` on macOS) on the file name in the UI. This will add a URL parameter with the files selected and you can share this URL with end users.

Old versions of `http-server` required only using `CTRL` (or `CMD) and clicking for highlighting, however, this was changed to avoid conflicts with the browser's default behavior of opening the file in a new tab.

Several file extensions include custom icons. More icons can be added provided the Icon Font in use, [Font Awesome](https://fontawesome.com/), supports the icon. If you see a generic file icon and you would like to have an extension to include a custom icon from Font Awesome, please open an issue.

#### Markdown rendering

When working as a directory listing tool, if the directory contains a `README.md`, `readme.md` or `index.md` file (either with `.md` or `.markdown` extension), it will be rendered as HTML and displayed on the directory listing page. You can choose to render the Markdown contents before or after the directory listing section: by default, it will render _after_ the directory listing. To render it _before_ you can use `--markdown-before-dir`.

The goal of this feature is to quickly provide the option for site operators to provide instructions to end users in the form of a document. You can combine the markdown rendering feature with the file highlighting feature to direct the end user's attention to specific files too, for example.

When using markdown, consider:

* CommonMark and GitHub Flavored Markdown are supported
* Mermaid diagrams are supported
* Raw HTML within markdown files is not supported for security considerations, instead, `<!-- raw HTML omitted -->` will be rendered
* Code fences are supported but they will not include syntax highlighting
* Headings will include anchors, and links to headings are supported
* Links to files within the directory being printed are also supported
* Images loaded from the directory in use are supported, as well as 3rd party images

#### Air-gapped environment

`http-server` is self-contained. Previous versions will load specific assets from the web, which made them unsuitable for environments with no internet access or in corporate environments. Starting from version `v2`, all assets are bundled with the binary, and the `http-server` will not load any external assets.

Behaviour-wise, this allows for custom icons in directory listing mode, as well as using the "Roboto" and "Roboto Mono" fonts to render the UI and potential markdown files. For these fonts, the following charsets are supported: `cyrillic`, `cyrillic-ext`, `greek`, `greek-ext`, `latin`, `latin-ext`, and `vietnamese`.

### Authentication

`http-server` supports two modes of authorizing access to its contents. On "directory listing" mode, only the directory contents are protected, while anything that's specific to `http-server`'s behaviour such as static assets like CSS, JavaScript or images are not protected. Everything served from the provided `--path` is protected.

#### Basic authentication

You can enable basic authentication by using the `--username` and `--password` flags. If both flags are provided, the server will require the provided username and password to access its contents.

#### JWT authentication

You can enable JWT authentication by using the `--jwt-key` flag. If the flag is provided, the server will require a valid JWT token to access its contents. The JWT token must be provided in the `Authorization` header, and it can be prefixed with `Bearer` followed by a space. Optionally, you can also provide the token via the `token` query parameter. Tokens are redacted when printed to the logs.

If the JWT token contains the claims `iss` (issuer, the issuing entity) and `sub` (subject, the entity the token is about, commonly used to provide a username), they will be printed to the application logs for auditing capabilities.

If the JWT token contains the claims `exp` (expiration time) and `nbf` (not before) and the `--ensure-unexpired-jwt` option is set, the token will be validated against the given times. If the token is expired (as in, the `exp` field is after the current time) or not yet valid (as in, the `nbf` field is before the current time), the request will be rejected.

### Configuration provided via config files

Besides CLI flags and environment variables, `http-server` supports also providing its settings via configuration files thanks to [Viper](https://github.com/spf13/viper). The configuration file must be located in the directory being used, and it must start with `.http-server.{ext}`, where `{ext}` can be any extension supported by Viper, either JSON, TOML, YAML, HCL, envfiles or Java properties config files.

The setting names match the CLI flags in their long form, without the single or double dash required by the CLI flags. For example, the `--path` flag can be set via the `path` setting in the configuration file.
