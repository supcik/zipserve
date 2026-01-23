# zipserve

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/supcik/zipserve)](https://goreportcard.com/report/github.com/supcik/zipserve)

A lightweight HTTP server that serves web content directly from a ZIP archive. Perfect for quickly viewing static websites offline or distributing self-contained web applications.

## Features

- 🚀 **Fast and lightweight** - Minimal dependencies, quick startup
- 📦 **Self-contained** - Serve complete websites from a single ZIP file
- 🌐 **Automatic browser** - Opens your default browser automatically (optional)
- 🔧 **URL prefix support** - Handle sites with path prefixes (e.g., GitHub Pages)
- 💻 **Cross-platform** - Works on Linux, macOS, and Windows

## Installation

### From Homebrew (macOS and Linux)

```bash
brew tap supcik/tap
brew install zipserve
```

### From Scoop (Windows)

```bash
scoop bucket add supcik https://github.com/supcik/scoop-bucket
scoop install zipserve
```

### From GitHub Releases

1. Go to the [releases page](https://github.com/supcik/zipserve/releases)
2. Download the appropriate binary for your platform:
   - **macOS**: `zipserve_darwin_amd64` (Intel) or `zipserve_darwin_arm64` (Apple Silicon)
   - **Linux**: `zipserve_linux_amd64` or `zipserve_linux_arm64`
   - **Windows**: `zipserve_windows_amd64.exe`
3. Make it executable (Linux/macOS):

   ```bash
   chmod +x zipserve_*
   ```

4. Move it to a directory in your PATH:

   ```bash
   sudo mv zipserve_* /usr/local/bin/zipserve
   ```

### From Source (Go required)

```bash
go install github.com/supcik/zipserve@latest
```

## Usage

### Basic Usage

```bash
zipserve mywebsite.wzip
```

This will start the server on port 8080 and automatically open your default browser.

### Command Line Options

```text
Usage:
  zipserve [flags] ZIPFILE

Flags:
  -d, --directory string   Directory to serve in the zip file. If not set, the directory
                           containing the .prefix file is used
  -h, --help               help for zipserve
  -p, --port int           Port Number (default 8080)
  -q, --prefix string      Path prefix. If not set, the prefix is read from the .prefix
                           file inside the zip file.
  -n, --skip-browser       Do not open the browser automatically
  -v, --verbose            Enable verbose logging
      --version            version for zipserve

ZIPFILE is the archive containing the web site (usually a .wzip file)
```

### Examples

```bash
# Serve on a different port
zipserve -p 3000 mywebsite.wzip

# Serve with a URL prefix
zipserve --prefix /my-project mywebsite.wzip

# Start server without opening browser
zipserve --skip-browser mywebsite.wzip
```

## How to build the archive

The archive is a simple zip file. You can typically
build it using the `zip` command :

```bash
cd public && zip -FS -r ../my-website.wzip . && cd ..
```

> **Note**: I chose the extension `.wzip` to make it possible
> to associate this extension with `zipserve` for easy opening of archives.

### URL Prefix Handling

Note that the archive is made at the root of the content,
but sometimes the URL of the site has a given prefix.
If the site is hosted on GitHub pages, the URL is
`https://<GROUP-NAME>.github.io/<PROJECT-NAME>/` and
the offline version will not work as expected if `/<PROJECT-NAME>/`
is not in the URL.

The `zipserve` program has the `--prefix` option that you
can use to provide the required prefix, but you can also
provide the correct prefix directly in the archive.

For this, just add the file `/.prefix` with one line
containing the correct prefix. For example :

```text
/my-project-name/
```

Build the zip with the file `/.prefix` and `zipserve`
will use the correct default.

## Use Cases

- **Offline documentation** - View generated documentation (Sphinx, MkDocs, etc.) without a web server
- **GitHub Pages preview** - Test your GitHub Pages site locally before deployment
- **Website distribution** - Share static websites as a single file
- **Educational purposes** - Distribute course materials or tutorials as self-contained packages
- **Quick demos** - Present static web projects without deployment

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Jacques Supcik - [@supcik](https://github.com/supcik)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI handling
- Uses [pkg/browser](https://github.com/pkg/browser) for cross-platform browser opening
