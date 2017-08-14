# Simple HTTP file server

[![Build Status](https://travis-ci.org/patrickdappollonio/http-server.svg?branch=master)](https://travis-ci.org/patrickdappollonio/http-server)
[![Docker](https://img.shields.io/docker/pulls/patrickdappollonio/docker-http-server.svg)](https://hub.docker.com/r/patrickdappollonio/docker-http-server/)

`http-server` is a simple binary to provide a static http server from a given folder. The
binary accepts a `-path` flag which can be pointed to a given folder to display it on a browser.

The application listen by default on port 5000. If the folder contains an `index.html` file
then the request will be redirected to the folder root and the `index.html` file will be displayed
instead of the file explorer (see screenshot below for the file explorer).

In different words, if you access `localhost:5000/index.html`, then the app will redirect to
`localhost:5000/` without the `index.html` part. All folders are enforced to be read by using a
trailing slash at the end.

There's also a Docker container you can use by mounting anything into the `/html` path. When served
from Docker, if no directory is mouted, it'll redirect by default here, to this repository.

## File explorer

![http-server file explorer](http://i.imgur.com/m8otA2i.png)

The `http-server` app includes a file explorer which can be useful to show some downloadable files.
In the screenshot example, I'm serving my `$HOME` directory and I navigated to `/Golang/src/github.com/patrickdappollonio/http-server`
(this repository) to showcase some of the features:

* File detection, based either on the file extension, file name or both.
* File size reported by the OS
* Modified date and permissions -- under the info icon next to the name
* Folders are always shown first, files below folders
* Path explorer on top: you can jump back and forth between different folders parent to the one you're visiting

## Container published to the docker registry

The docker container [is published in the public Docker Registry](https://hub.docker.com/r/patrickdappollonio/docker-http-server/)
under `patrickdappollonio/docker-http-server`, you can pull it by executing:

```bash
docker pull patrickdappollonio/docker-http-server
```

## Use it with Docker standalone

Run the container, preferably in detached mode (by passing `-d`), exposing either
a random port with `-P` (uppercase "P"), or an actual mapping, with `-p 5000:5000`,
and mount the contents you want to show into the `/html` path.

```bash
# To get a random port from the ones available
docker run -d -P -v $(pwd):/html patrickdappollonio/docker-http-server

# To get a predefined port (in this case, 8080)
docker run -d -p 8080:5000 -v $(pwd):/html patrickdappollonio/docker-http-server
```

## Use it with Docker Compose

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
