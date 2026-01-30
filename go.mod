module lark

go 1.21

require (
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.9
)

replace github.com/cpuguy83/go-md2man/v2 => ./third_party/go-md2man
replace github.com/inconshreveable/mousetrap => ./third_party/mousetrap
replace go.yaml.in/yaml/v3 => ./third_party/yaml
