# zipserve

This program is an HTTP server taking its content
from a zip file.

## Usage

```text
zipserve [options] ZIPFILE

The options are:
      --port int        Port Number (default 8080)
      --prefix string   Path prefix

ZIPFILE is the archive containing the web site (usually a .wzip file)
```

## How to build the archive

The archive is a simple zip file. You can typically
build it using the `zip` command :

```bash
cd public && zip -FS -r ../my1-website.wzip . && cd ..
```

Note that I chose the extension `.wzip` to make is possible
to associate the extension with `.wzip` with `zipserve`

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

```
/my-project-name/
```

Build the zip with the file `/.prefix` and `zipserve`
will use the correct default.
