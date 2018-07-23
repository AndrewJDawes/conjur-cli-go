package main

import (
	"github.com/urfave/cli"
)

// AuthnCommands contains the definitions of the commands related to authentication.
var AuthnCommands = cli.Command{
	Name:  "authn",
	Usage: "Login and logout",
	Subcommands: []cli.Command{
		{
			Name:  "login",
			Usage: "log in to Conjur",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:  "whoami",
			Usage: "show current Conjur account and username",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	},
}
