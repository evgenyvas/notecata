.. index::
   single: api

API
===

Server **notecatad** allows http REST requests to manage your notes storage

Create note
-----------

POST {URL}/api/notes/{PATH}

Example request:

POST http://localhost:8080/api/notes/ddocs.md

::

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

::

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

Get notes list
--------------

GET {URL}/api/noteslist/{PATH}

Example request:

GET http://localhost:8080/api/noteslist/docs/

Example output:

::

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

Delete note
-----------

DELETE {URL}/api/notes/{PATH}

Example request:

DELETE http://localhost:8080/api/notes/doc.md

Example output:

::

  {
      "status": 200,
      "message": "Note doc.md deleted successfully"
  }
