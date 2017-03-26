package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const verbosity = 0

func main() {
	fmt.Println("Launching go-tcp daemon")
	tap, err := initTAP()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init TAP dev: %s", err)
		os.Exit(1)
	}
	defer tap.Close()
	done := make(chan bool)
	in := make(chan *EthernetFrame)
	go handleSignals(done)
	go tap.ReceiveLoop(in, done)
	for f := range in {
		fmt.Printf("Ethernet frame: %s\n", f)
		if err := handleFrame(f); err != nil {
			fmt.Fprintf(os.Stderr, "failed to handle frame: %s", err)
		}
	}
	fmt.Printf("Done.\n")
}

func handleSignals(done chan bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Printf("Starting shutdown\n")
	done <- true
}
