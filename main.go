package main

import (
	"fmt"
	"log"
	"os"

	"github.com/theobitoproject/kankuro/pkg/source"
)

type logger struct{}

func (l *logger) Write(p []byte) (n int, err error) {
	msg := string(p)
	fmt.Println(msg)

	return len(msg), nil
}

func main() {
	hsrc := NewRandomAPISource("https://random-data-api.com/api/v2")
	runner := source.NewSafeSourceRunner(hsrc, &logger{}, os.Args)
	// runner := source.NewSafeSourceRunner(hsrc, os.Stdout, os.Args)
	err := runner.Start()
	if err != nil {
		log.Fatal(err)
	}
}
