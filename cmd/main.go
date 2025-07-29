package main

import (
	"BookStore/internal/common/config"
	"BookStore/internal/control/app"
	"log"
)

func main() {
	var cfg config.BaseConfig

	if err := app.Run(&cfg); err != nil {
		log.Fatal(err)
	}

}
