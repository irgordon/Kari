package main

import (
	"log"

	"github.com/kari/brain/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
