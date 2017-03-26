package main

import (
	"fmt"
	"net"
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
		parseEthernetFrame(b)
	}
}

type EthernetFrame struct {
	Dmac net.HardwareAddr
	Smac net.HardwareAddr
}

func (f *EthernetFrame) String() string {
	return fmt.Sprintf("src %s dest %s", f.Smac, f.Dmac)
}

func parseEthernetFrame(b []byte) {
	frame := &EthernetFrame{
		Dmac: net.HardwareAddr(b[:6:6]),
		Smac: net.HardwareAddr(b[6:12:12]),
	}
	fmt.Printf("Parsed Ethernet frame: %s\n", frame)
}
