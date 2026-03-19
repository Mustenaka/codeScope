package main

import (
	"log"

	"codescope/server/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
