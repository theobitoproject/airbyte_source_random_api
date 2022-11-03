package main

import (
	"log"
	"os"

	"github.com/theobitoproject/kankuro/pkg/source"
)

func main() {
	src := newSourceRandomApi("https://random-data-api.com/api/v2")
	runner := source.NewSafeSourceRunner(src, os.Stdout, os.Args)
	err := runner.Start()
	if err != nil {
		log.Fatal(err)
	}
}
