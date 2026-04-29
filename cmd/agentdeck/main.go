package main

import (
	"log"
	"os"

	"github.com/ajorgensen/agentdeck/internal/agentdeck"
)

func main() {
	if err := agentdeck.NewApp().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
