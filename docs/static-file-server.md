# Static file server

The core nature of `http-server` is to be a static file server. You can serve any folder in the node where `http-server` is running. **None of the files are hidden**, which means if the user that's executing `http-server` can see them, then they will be listed. The exceptions are:

* The `.http-server.yaml` configuration file, which is removed from view and direct access since it may contain sensitive information.
* The `_redirects` file used for redirection rules.
* TLS certificate and key files, if they reside inside the served directory and TLS is active. These are hidden from listings and blocked from direct download. See the [TLS documentation](tls.md) for details.

The files served are type-hinted and their `Content-Type` header set through this method. The server also supports `Accept-Ranges` header, meaning you can perform partial requests for bigger files and ensure it's possible to download them in chunks if needed.
