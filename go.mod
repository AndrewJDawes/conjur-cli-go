module github.com/cyberark/conjur-cli-go

go 1.19

// Use the replace below for local development with conjur-api-go
// replace github.com/cyberark/conjur-api-go => path/to/conjur-api-go

require (
	github.com/cyberark/conjur-api-go v0.10.2-0.20220921161402-2084bc6d5e4a // Run "go get github.com/cyberark/conjur-api-go@for-cli" to update
	github.com/manifoldco/promptui v0.9.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20211214234402-4825e8c3871d // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
