module github.com/tibco/flogo-tdv/src/app/TDV

go 1.20

require (
	github.com/alexbrainman/odbc v0.0.0-20211220213544-9c9a2e61c5e2
	github.com/project-flogo/core v1.6.6
	github.com/tibco/wi-contrib v1.0.0-rc1
)

replace github.com/alexbrainman/odbc => ./vendorUpdates/github.com/alexbrainman/odbc

require (
	github.com/TIBCOSoftware/flogo-lib v0.5.8 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)
