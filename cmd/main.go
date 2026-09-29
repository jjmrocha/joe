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
	ctx := context.Background()

	if err := setup.BuildIfNeeded(ctx); err != nil {
		log.Fatal(err)
	}

	profile := config.DefaultProfile
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
