package cli

import (
	"errors"

	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// InitCommands contains the definition of the command for initialization.
var InitCommands = func(api cmd.ConjurClient, fs afero.Fs) []cli.Command {
	return []cli.Command{
		{
			Name:  "init",
			Usage: "Initialize the Conjur configuration",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "account, a",
					Usage: "Conjur organization account name",
				},
				cli.StringFlag{
					Name:  "certificate, c",
					Usage: "Conjur SSL certificate (will be obtained from host unless provided by this option)",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "File to write the configuration to",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "Force overwrite of existing files",
				},
				cli.StringFlag{
					Name:  "url, u",
					Usage: "URL of the Conjur service",
				},
			},
			Action: func(c *cli.Context) error {
				return errors.New("not yet implemented")
			},
		},
	}
}
