package main

import (
	"log"
	"os"

	"github.com/theobitoproject/kankuro/source"
)

func main() {
	hsrc := NewRandomAPISource("https://random-data-api.com/api/v2")
	runner := source.NewSourceRunner(hsrc, os.Stdout)
	err := runner.Start()
	if err != nil {
		log.Fatal(err)
	}
}
