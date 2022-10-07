package main

import (
	"log"
	"os"
	"time"

	"github.com/rog-golang-buddies/rmx/internal/commands"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	if err := initCLI().Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func initCLI() *cli.App {
	c := &cli.App{
		EnableBashCompletion: true,
		Name:                 "rmx",
		Usage:                "RapidMidiEx Server CLI",
		Version:              "v0.0.1",
		Compiled:             time.Now().UTC(),
		Action: func(*cli.Context) error {
			return nil
		},
		Before: altsrc.InitInputSourceWithContext(
			commands.Flags,
			altsrc.NewYamlSourceFromFlagFunc("load"),
		),
		Commands: commands.Commands,
	}

	return c
}
