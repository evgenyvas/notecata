.. index::
   single: install

Install
=======

Compile both cli and server

::

  go build cmd/cli/notecata.go
  go build cmd/server/notecatad.go

edit configuration file `config.json`:

`root` - path to directory to store notes if storage type **file** is selected
`port` - which port will be used by **notecatad** server
`storage` - type of storage for notes, now support only **file**, which means in OS filesystem

That's all, you can now use CLI interface **notecata**.

If you want to use WEB or REST interface, start **notecatad** server:

::

  ./notecatad

If **port** paramenter in config file equals `8080`, WEB interface will be available by address http://localhost:8080

Or you can install systemd service, use example in `support/systemd/notecata.service`

