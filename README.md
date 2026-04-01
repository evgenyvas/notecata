NoteCata
========

Notes Catalogue

Install
=======

Compile both cli and server

```
  go build cmd/cli/notecata.go
  go build cmd/server/notecatad.go
```

edit configuration file `config.json`:

- root - path to directory to store notes if storage type **file** is selected
- port - which port will be used by **notecatad** server
- storage - type of storage for notes, now support only **file**, which means in OS filesystem

That's all, you can now use CLI interface **notecata**.

If you want to use WEB or REST interface, start **notecatad** server:

```
  ./notecatad
```

If **port** paramenter in config file equals `8080`, WEB interface will be available by address http://localhost:8080

Or you can install systemd service, use example in `support/systemd/notecata.service`

CLI
===

For command-line interface used **notecata** tool

Create note
-----------

To create note you must pass data to stdin and set path to save it.

```
  ./notecata doc.md < ~/mynote.md
```

Or for path with directories:

```
  ./notecata docs/test/doc.md < ~/mynote.md
```

Parameter `--meta, -meta, -m` allows to make an interactive ask for note metadata

```
  ./notecata -m docs/test/doc.md < ~/mynote.md
```

If you will use already existed path, old note will be rewritten.

If you will set path to directory, tool will ask for filename

```
  ./notecata docs < ~/mynote.md
```

For root directory set **/** as argument

```
  ./notecata / < ~/mynote.md
```

View notes
----------

To view note content just set path to it:

```
  ./notecata docs.md
```

Data will be written to stdout, so you can use external tool for view it:

```
  ./notecata docs.md | nvim
```

For directory path tool will print it's content

```
  ./notecata docs
  ./notecata /
```

Delete notes
------------

To delete note or directory pass parameter `--del, -del, -d` and set path:

```
  ./notecata -d docs/test/doc.md
  ./notecata -d docs/test
```

If path equals directory, it will be deleted with all content inside it.

API
===

Server **notecatad** allows http REST requests to manage your notes storage

Create note
-----------

POST {URL}/api/notes/{PATH}

Example request:

POST http://localhost:8080/api/notes/ddocs.md

```
  {
      "format": "Markdown",
      "title": "api test",
      "tags": [
          "go",
          "yaml",
          "json"
      ],
      "content": "Note content"
  }
```

* format - now support only **Markdown**
* title - string with metadata info
* tags - array with strings, metadata tags
* content - text with note content, put your markdown here

Get single note
---------------

GET {URL}/api/notes/{PATH}

Example request:

GET http://localhost:8080/api/notes/doc.md

Example output:

```
  {
      "status": 200,
      "message": "ok",
      "path": "docs.md",
      "format": "Markdown",
      "date": "2026-04-01 20:13:02",
      "title": "The Go Programming Language",
      "tags": [
          "go",
          "golang"
      ],
      "content": "My note content",
      "type": "note"
  }
```

Get notes list
--------------

GET {URL}/api/noteslist/{PATH}

Example request:

GET http://localhost:8080/api/noteslist/docs/

Example output:

```
  {
      "status": 200,
      "message": "ok",
      "notes": [
          {
              "path": "docs.md",
              "format": "Markdown",
              "date": "2026-04-01 20:13:02",
              "title": "The Go Programming Language",
              "tags": [
                  "go",
                  "golang"
              ],
              "content": "My note content",
              "type": "note"
          },
          {
              "path": "/",
              "name": "test",
              "type": "directory"
          }
      ]
  }
```

Delete note
-----------

DELETE {URL}/api/notes/{PATH}

Example request:

DELETE http://localhost:8080/api/notes/doc.md

Example output:

```
  {
      "status": 200,
      "message": "Note doc.md deleted successfully"
  }
```
