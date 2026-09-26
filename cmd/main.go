package main

import (
	"context"
	"log"
	"os"

	"github.com/jjmrocha/joe/internal/config"
	"github.com/jjmrocha/joe/internal/engine"
	"github.com/jjmrocha/joe/internal/setup"
)

func main() {
	if err := setup.BuildIfNeeded(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	profile := "default"
	if len(os.Args) > 1 {
		profile = os.Args[1]
	}

	cfg, err := config.Load(profile)
	if err != nil {
		log.Fatal(err)
	}

	if err := engine.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
