package main

import (
	"context"
	"log"

	"github.com/jjmrocha/joe/internal/engine"
)

func main() {
	ctx := context.Background()

	if err := engine.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
