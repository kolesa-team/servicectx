module github.com/kolesa-team/servicectx/otel

go 1.17

replace github.com/kolesa-team/servicectx => ../

require (
	github.com/kolesa-team/servicectx v0.1.1-0.20220311063942-2e3c1782177a
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.4.1
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/trace v1.4.1 // indirect
	gopkg.in/yaml.v3 v3.0.0 // indirect
)
