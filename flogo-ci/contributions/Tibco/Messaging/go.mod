module github.com/tibco/flogo-messaging/src/app/Messaging

go 1.20

replace github.com/TIBCOSoftware/eftl => ./eftl

require (
	github.com/TIBCOSoftware/eftl v0.0.0-00010101000000-000000000000
	github.com/project-flogo/core v1.6.6
	github.com/stretchr/testify v1.8.4
	github.com/tibco/wi-contrib v1.0.0-rc1
)

require (
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/zerolog v1.26.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
