# CORS support

`http-server` supports CORS, which means you can use it to serve files to other domains. This is done by setting the `Access-Control-Allow-Origin` header to `*`, which means any domain can access the files.

The server also supports `HEAD` requests, however, considering the `http-server` only supports `GET` requests, they should be unnecessary for a full CORS experience since preflight requests are only made for non-`GET` requests only.

CORS requests work both in file server mode as well as directory listing mode.
