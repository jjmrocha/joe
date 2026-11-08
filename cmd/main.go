package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jjmrocha/joe/internal/cli"
	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/engine"
	"github.com/jjmrocha/joe/internal/setup"
)

func main() {
	ctx := context.Background()

	args, err := cli.Parse(os.Args[1:])
	if err != nil {
		cli.Usage(os.Stdout)

		if errors.Is(err, cli.ErrHelp) {
			os.Exit(0)
		}

		os.Exit(1)
	}

	if err = setup.BuildIfNeeded(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load(args.Profile)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Run(ctx, cfg, args.SessionID, args.Debug); err != nil {
		log.Fatal(err)
	}
}
