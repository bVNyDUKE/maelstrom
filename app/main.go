package main

import (
	"log"
	"os"
)

func main() {
	node := NewNode()

	if err := node.Run(); err != nil {
		log.Printf("Error running node: %s", err)
		os.Exit(1)
	}
}
