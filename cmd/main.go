package main

import (
	"context"
	"log"
	"os"

	"github.com/jjmrocha/joe/internal/engine"
)

func main() {
	ctx := context.Background()

	profile := "default"
	if len(os.Args) > 1 {
		profile = os.Args[1]
	}

	if err := engine.Run(ctx, profile); err != nil {
		log.Fatal(err)
	}
}
