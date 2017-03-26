package main

import (
	"encoding/hex"
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

type EtherType [2]byte

var (
	ARP = EtherType{0x08, 0x06}
)

func (e EtherType) String() string {
	switch e {
	case ARP:
		return "ARP"
	}
	out := make([]byte, 4)
	hex.Encode(out, e[:])
	return fmt.Sprintf("%s", string(out))
}

type EthernetFrame struct {
	Dmac      net.HardwareAddr
	Smac      net.HardwareAddr
	EtherType EtherType
	Payload   []byte
}

func (f *EthernetFrame) String() string {
	return fmt.Sprintf("src %s dest %s type %s payload size %d", f.Smac, f.Dmac, f.EtherType, len(f.Payload))
}

func parseEthernetFrame(b []byte) {
	frame := &EthernetFrame{
		Dmac:      net.HardwareAddr(b[:6:6]),
		Smac:      net.HardwareAddr(b[6:12:12]),
		EtherType: EtherType{b[12], b[13]},
		Payload:   b[14:],
	}
	fmt.Printf("Parsed Ethernet frame: %s.\n", frame)
}
