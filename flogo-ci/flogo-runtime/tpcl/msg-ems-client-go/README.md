# TIBCO EMS Go client library

Godoc coming soon; for now, see the [samples](samples) directory for how to build sample applications.

## Running Unit Tests

```
PKG_CONFIG_PATH=/path/to/msg-ems-client-go/pkg-config/{platform} CGO_ENABLED=1 GOARCH=amd64 go test github.com/tibco/msg-ems-client-go/tibems
```

## API Documentation

To view the current work-in-progress API doc, simply install `pkgsite`:

```
go install golang.org/x/pkgsite/cmd/pkgsite@latest
```

and run it:

```
cd /path/to/msg-ems-client-go
$GOPATH/bin/pkgsite -http localhost:6060
```

API documentation should now be available at [http://localhost:6060/github.com/tibco/msg-ems-client-go@v0.0.0/tibems](http://localhost:6060/github.com/tibco/msg-ems-client-go@v0.0.0/tibems) .

Note that the Go documentation is currently very limited; please consult the existing EMS C API documentation to fill in any gaps.
