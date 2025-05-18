# Force Download Extensions

The `--force-download-extensions` flag allows you to specify file extensions that should always be downloaded when accessed through the server, rather than being displayed in the browser.

## Usage

You can specify multiple extensions in one of these ways:

1. Comma-separated list:
```bash
http-server --force-download-extensions jpg,pdf,zip
```

2. Multiple flag occurrences:
```bash
http-server --force-download-extensions jpg --force-download-extensions pdf
```

3. In a configuration file (.http-server.yaml):
```yaml
force-download-extensions:
  - jpg
  - pdf
  - zip
```

4. As environment variables:
```bash
export FILE_SERVER_FORCE_DOWNLOAD_EXTENSIONS=jpg,pdf,zip
```

## How it works

When a client requests a file with an extension that matches one in the list, the server will add a `Content-Disposition: attachment; filename="filename.ext"` header to the response. This tells the browser to download the file rather than attempting to display it.

This is useful for:
- Image files that you want users to download instead of view in the browser
- PDF documents that should be saved locally rather than opened in a browser
- Media files or any other content you want to force as a download

## Limitations

- **Markdown in directory listing**: When using the directory listing feature, Markdown files (like `README.md`) that are rendered as part of the page won't be affected by this flag, since they're processed before the file serving logic is applied.
- **Direct markdown links**: However, if a user directly accesses a Markdown file via URL (not through directory listing), the force-download setting will apply if .md/.markdown is in the list of extensions.

## Example

If you configure the server with `--force-download-extensions jpg,png`, any requests to files ending in `.jpg` or `.png` will be downloaded automatically instead of being displayed in the browser.
