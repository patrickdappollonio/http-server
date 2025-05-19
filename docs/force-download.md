# Force Download Extensions

The `--force-download-extensions` flag allows you to specify file extensions or exact filenames that should always be downloaded when accessed through the server, rather than being displayed in the browser.

## Usage

You can specify multiple extensions or filenames in one of these ways:

1. Comma-separated list:
```bash
# Simple extensions
http-server --force-download-extensions jpg,pdf,zip

# Compound extensions
http-server --force-download-extensions min.js,min.css

# Exact filenames (case-insensitive)
http-server --force-download-extensions Dockerfile,.env
```

2. Multiple flag occurrences:
```bash
http-server --force-download-extensions jpg --force-download-extensions min.js
```

3. In a configuration file (.http-server.yaml):
```yaml
force-download-extensions:
  - jpg
  - min.js
  - Dockerfile
  - .env
```

4. As environment variables:
```bash
export FILE_SERVER_FORCE_DOWNLOAD_EXTENSIONS="jpg,min.js,Dockerfile,.env"
```

## How it works

When a client requests a file, the server checks if the file's extension or full filename matches any in the force download list (`--force-download-extensions`). If a match is found, the server adds a `Content-Disposition: attachment; filename="filename.ext"` header to the response. This tells the browser to download the file rather than attempting to display it.

### Supported patterns

- **Simple extensions**: `.jpg`, `.pdf`, `.zip`
  - Matches any file with that extension (case-insensitive)
  - Example: `image.jpg`, `document.PDF`

- **Compound extensions**: `.min.js`, `.bundle.css`
  - Matches the exact extension sequence (case-insensitive)
  - Example: `app.min.js`, `styles.MIN.CSS`

- **Exact filenames**: `Dockerfile`, `.env`
  - Matches files with exactly the same name (case-insensitive)
  - Example: `Dockerfile`, `.env`, `dockerfile`

## Use cases

- Force download of image files instead of displaying them in the browser
- Ensure PDF documents are saved locally rather than opened in the browser
- Handle minified or bundled assets with compound extensions
- Control download behavior for files without extensions like `Dockerfile`

## Limitations

- **Markdown in directory listing**: When using the directory listing feature, Markdown files (like `README.md`) that are rendered as part of the page won't be affected by this flag, since they're processed before the file serving logic is applied.
- **Direct markdown links**: If a user directly accesses a Markdown file via URL (not through directory listing), the force-download setting will apply if `.md` or `.markdown` is in the list of extensions.
- **Case sensitivity**: All matches are case-insensitive for consistency across different operating systems.
- **Skip list precedence**: The skip list (`--skip-force-download-files`) takes precedence over the force-download list. If a file matches both lists, it will be served normally (not force-downloaded).

## Examples

1. Force download of common media files:
   ```bash
   http-server --force-download-extensions jpg,png,gif,mp4,pdf
   ```

2. Handle minified JavaScript and CSS:
   ```bash
   http-server --force-download-extensions min.js,min.css
   ```

3. Force download CSS files but exclude specific ones:
   ```bash
   http-server --force-download-extensions css --skip-force-download-files styles.css,themes/dark.css
   ```

4. Complex configuration in YAML:
   ```yaml
   force-download-extensions:
     - css
     - js
     - png
   skip-force-download-files:
     - styles.css
     - scripts/main.js
     - assets/logo.png
   ```
