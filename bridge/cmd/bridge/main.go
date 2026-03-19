package main

import (
	"log"

	"codescope/bridge/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
