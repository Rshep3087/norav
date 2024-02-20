package main

import (
	"log"

	"github.com/rshep3087/norav/cmd"
)

type statusMsg map[string]int

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
