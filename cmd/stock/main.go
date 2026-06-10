package main

import (
	"context"
	"fmt"
	"os"

	"stock-cli/pkg/cmd"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := cmd.Command.Run(context.Background(), os.Args); err != nil {
		exitCode := 1
		if exitErr, ok := err.(cli.ExitCoder); ok {
			exitCode = exitErr.ExitCode()
		}
		if cmd.CommandErrorBuffer.Len() > 0 {
			_, _ = os.Stderr.Write(cmd.CommandErrorBuffer.Bytes())
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
		os.Exit(exitCode)
	}
}
