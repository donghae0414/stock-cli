package cmd

import (
	"bytes"

	"github.com/urfave/cli/v3"
)

const Version = "0.1.0"

var (
	Command            *cli.Command
	CommandErrorBuffer bytes.Buffer
)

func init() {
	Command = &cli.Command{
		Name:      "stock",
		Usage:     "CLI for stock trading APIs",
		Suggest:   true,
		Version:   Version,
		ErrWriter: &CommandErrorBuffer,
		Commands: []*cli.Command{
			&accountsCmd,
			&configCmd,
		},
	}
}
