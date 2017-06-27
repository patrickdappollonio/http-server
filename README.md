# Docker container for a HTTP file server

[![Build Status](https://travis-ci.org/patrickdappollonio/docker-http-server.svg?branch=master)](https://travis-ci.org/patrickdappollonio/docker-http-server)
[![Docker Automated buil](https://img.shields.io/docker/automated/patrickdappollonio/docker-http-server.svg)](https://hub.docker.com/r/patrickdappollonio/docker-http-server/)

This docker container is just a simple HTTP file server. It will serve a simple file server
which will show either the contents of the `/html` folder or the mounted volume contents.

By default, the container will listen in port `5000` accepting any incoming request. If no volume
is passed to be mount, then [the default html currently available](html/index.html) will redirect
you here, to this docs.

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
