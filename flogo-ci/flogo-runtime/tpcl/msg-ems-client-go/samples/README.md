To build the samples in this directory:

`
CGO_ENABLED=1 GOARCH=amd64 PKG_CONFIG_PATH=/full/path/to/msg-ems-client-go/pkg-config/{platform} go build samples/msg-producer/msg-producer.go
`

`
CGO_ENABLED=1 GOARCH=amd64 PKG_CONFIG_PATH=/full/path/to/msg-ems-client-go/pkg-config/{platform} go build samples/msg-consumer/msg-consumer.go
`

etc.

You must have TIBCO EMS installed at the location specified in the `/full/path/to/msg-ems-client-go/pkg-config/{platform}/ems.pc` file you use to build.