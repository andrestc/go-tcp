package main

import (
	"fmt"
	"os"
)

const verbosity = 0

func main() {
	fmt.Println("Launching go-tcp daemon")
	tap, err := initTAP()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init TAP dev: %s", err)
		os.Exit(1)
	}
	ch := make(chan []byte)
	go tap.Loop(ch)
	for b := range ch {
		fmt.Printf("%d\n", b)
	}
}
