.. index::
   single: cli

CLI
===

For command-line interface used **notecata** tool

Create note
-----------

To create note you must pass data to stdin and set path to save it.

::

  ./notecata doc.md < ~/mynote.md

Or for path with directories:

::

  ./notecata docs/test/doc.md < ~/mynote.md

Parameter `--meta, -meta, -m` allows to make an interactive ask for note metadata

::

  ./notecata -m docs/test/doc.md < ~/mynote.md

If you will use already existed path, old note will be rewritten.

If you will set path to directory, tool will ask for filename

::

  ./notecata docs < ~/mynote.md

For root directory set **/** as argument

::

  ./notecata / < ~/mynote.md

View notes
----------

To view note content just set path to it:

::

  ./notecata docs.md

Data will be written to stdout, so you can use external tool for view it:

::

  ./notecata docs.md | nvim

For directory path tool will print it's content

::

  ./notecata docs
  ./notecata /

Delete notes
------------

To delete note or directory pass parameter `--del, -del, -d` and set path:

::

  ./notecata -d docs/test/doc.md
  ./notecata -d docs/test

If path equals directory, it will be deleted with all content inside it.
