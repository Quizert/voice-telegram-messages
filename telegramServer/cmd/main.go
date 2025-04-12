package main

import (
	"kursach/app"
	"log"
)

func main() {
	app := app.NewApp()
	err := app.Init()
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	app.Start()
}
