module github.com/tibco/wi-salesforce/src/app/Salesforce

go 1.20

replace github.com/zph/bayeux => ./bayeux

require (
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/project-flogo/core v1.6.6
	github.com/stretchr/testify v1.8.4
	github.com/tibco/wi-contrib v1.0.0-rc1
	github.com/zph/bayeux v1.0.0
)

require (
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
