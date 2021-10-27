# Simple HTTP file server

[![Docker](https://img.shields.io/docker/pulls/patrickdappollonio/docker-http-server.svg)](https://hub.docker.com/r/patrickdappollonio/docker-http-server/)
[![Github Downloads](https://img.shields.io/github/downloads/patrickdappollonio/http-server/total?color=orange&label=github%20downloads)](https://github.com/patrickdappollonio/http-server/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/patrickdappollonio/http-server)](https://goreportcard.com/report/github.com/patrickdappollonio/http-server)

`http-server` is a simple binary to provide a static http server from a given folder. The binary accepts a `-path` flag which can be pointed to a given folder to display it on a browser.

The application listen by default on port 5000. If the folder contains an `index.html` file then the request will be redirected to the folder root and the `index.html` file will be displayed instead of the file explorer (see screenshot below for the file explorer).

In different words, if you access `localhost:5000/index.html`, then the app will redirect to `localhost:5000/` without the `index.html` part. All folders are enforced to be read by using a trailing slash at the end.

Files are served using the Go standard HTTP file server, which handles Range requests properly, sets the MIME type, and handles `If-Match`, `If-Unmodified-Since`, `If-None-Match`, `If-Modified-Since`, and `If-Range` requests. The MIME-type is based off the file extension or by reading the first 512 bytes of the file and trying to guess the file type based off some rules. If nothing matches then the `application/octet-stream` header is set as the content type.

Also, based off the modified date from the file being browsed, `http-server` includes a Last-Modified header in the response. If the request includes an `If-Modified-Since` header, `http-server` will use the modified time to decide whether the content needs to be sent at all.

If there's a proper `w`'s `ETag` -- formatted based on RFC 7232, section 2.3 -- then the server will use it to appropriately handle `If-Match`, `If-None-Match` or `If-Range`.

There's also a Docker container you can use by mounting anything into the `/html` path. When served from Docker, if no directory is mouted, it'll redirect by default here, to this repository.

### File explorer

![http-server file explorer](http://i.imgur.com/m8otA2i.png)

The `http-server` app includes a file explorer which can be useful to show some downloadable files. In the screenshot example, I'm serving my `$HOME` directory and I navigated to `/Golang/src/github.com/patrickdappollonio/http-server` (this repository) to showcase some of the features:

* **File type detection**, based either on the file extension, file name or both.
* **File size** reported by the OS
* **Modified date and permissions** -- under the info icon next to the name
* **Folders are always shown first**, files below folders
* **Path explorer on top**: you can jump back and forth between different folders parent to the one you're visiting
* **CTRL-click to select one or many folders or files**. The selected folders / files will be highlighted (useful when you want to direct the attention to one or more files but still show the rest of them).
* **One-to-one mapping of folders to URL path**: if you change manually the URL, the application will list the files on that path.

### Configurable options

All of the following options set different settings regarding how the `http-server` behaves:

* `$FILE_SERVER_PATH` or `--path`: Allows you to set a different path than the default one, which is `/html`
* `$FILE_SERVER_TITLE`: Customize the title to whatever you want. By default the UI title is "HTTP File Server" and the HTML page title is "Browsing: /path/to/file".
  This variable sets both to the same value.
* `$FILE_SERVER_COLOR_SET`: Choose from a pretty big set of different colors schemes for your UI. The supported list of colors
  [is available in the MDL website](https://getmdl.io/customize/index.html). Choose from the selector and then you'll get an URL
  which is always ending on `material.$COLOR.min.css`. Pick the `$COLOR` part and use it here. By default I pick `indigo-red` for you.
* `$FILE_SERVER_PATH_PREFIX` or `--pathprefix`: Allows you to set a different path prefix than the default one, which is `/`. Useful when this needs to be hosted under a subdirectory rather than in the domain root.
* `$FILE_SERVER_PORT`, `$PORT` or `--port`: Allows you to set a different port than the default one, which is `5000`.
* `$FILE_SERVER_TITLE` or `$PAGE_TITLE`: Allows you to set a different page title than the default one, which is `HTTP File Server`.
* `$FILE_SERVER_USERNAME` or `$HTTP_USER`: Sets the HTTP username for any request to the server. In order for this to work, the HTTP password must be set as well.
* `$HTTP_SERVER_PASSWORD` or `$HTTP_PASS`: Sets the HTTP password for any request to the server. In order for this to work, the HTTP username must be set as well.
* `$FILE_SERVER_HIDE_LINKS` or `$HIDE_LINKS`: If set to any non-empty value, it hides the top links linking to Github and the Docker Hub.
* `$FILE_SERVER_BANNER`, `$BANNER` or `--banner`: Allows you to set a custom banner to be displayed on the top of the page. **Warning:** any contents passed here will be rendered as-is. If HTML code is passed, it will print that inside the banner block.

## Usage with Docker

#### Container published to the docker registry

The docker container [is published in the public Docker Registry](https://hub.docker.com/r/patrickdappollonio/docker-http-server/)
under `patrickdappollonio/docker-http-server`, you can pull it by executing:

```bash
docker pull patrickdappollonio/docker-http-server
```

#### Use it with Docker standalone

Run the container, preferably in detached mode (by passing `-d`), exposing either
a random port with `-P` (uppercase "P"), or an actual mapping, with `-p 5000:5000`,
and mount the contents you want to show into the `/html` path.

```bash
# To get a random port from the ones available
docker run -d -P -v $(pwd):/html patrickdappollonio/docker-http-server

# To get a predefined port (in this case, 8080)
docker run -d -p 8080:5000 -v $(pwd):/html patrickdappollonio/docker-http-server
```

#### Use it with Docker Compose

To use it with Docker Compose you need to create a `docker-compose.yaml` and paste inside
the content below. Make sure to change the port mapping `5000:5000` and the folder you want
to serve (currently, `./html:/html`).

```yaml
version: '2'

services:
  http-server:
    image: patrickdappollonio/docker-http-server
    ports:
      - 5000:5000
    volumes:
      - ./html:/html
    restart: always
```
