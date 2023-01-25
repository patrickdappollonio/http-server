# Authentication support

`http-server` supports two modes of authorizing access to its contents. On [directory listing](directory-listing.md) mode, only the directory contents are protected, while anything specific to `http-server`'s behaviour such as static assets like CSS, JavaScript or images are not protected. Everything served from the provided `--path` is protected.

### Plain username and password

You can enable basic authentication by using the `--username` and `--password` flags. If both flags are provided, the server will require the provided username and password to access its contents.

This is the simplest form of authentication. The username and password are sent in plain text over the network if you are not serving `http-server` via HTTPS. As such, it's not recommended for production use. If you still decide to use it, use a strong password.

### JWT authentication

This is a more secure form of authentication. It uses [JSON Web Tokens](https://jwt.io/) to authenticate requests. The JWT token must be provided in the `Authorization` header or via the `token` querystring parameter. If passed via the header, it must be prefixed with `Bearer` followed by a space.

The way JWT tokens authenticate is, provided you pass the signing key to `http-server`, the server will validate the token's signature and, if the token is signed with the provided key, it will be considered valid. It will be invalid and rejected instead. As long as the signing key is not compromised, the token should be safe.

Additionally, you can enable time validation for JWT claims `exp` and `nbf` by using the `--ensure-unexpired-jwt` flag. This will ensure that the token is not expired and that it's not used before its `nbf` claim. Use this to your advantage to create short-lived tokens that expire after a certain amount of time, so if they were to be compromised, they would be useless after they expire.

Finally, if the JWT token contains the claims `iss` (issuer, the issuing entity) and `sub` (subject, the entity the token is about, commonly used to provide a username), they will be printed to the application logs for auditing capabilities. That way, you can track users of your application and who accessed what.
