package main

import (
	"log"
	"os"

	"github.com/theobitoproject/kankuro/pkg/source"
)

func main() {
	hsrc := NewRandomAPISource("https://random-data-api.com/api/v2")
	runner := source.NewSafeSourceRunner(hsrc, os.Stdout, os.Args)
	err := runner.Start()
	if err != nil {
		log.Fatal(err)
	}
}
