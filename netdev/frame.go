package netdev

import (
	"encoding/hex"
	"fmt"
	"net"
)

var ARP = EtherType{0x08, 0x06}

type EtherType [2]byte

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

func newFrame(b []byte) *EthernetFrame {
	return &EthernetFrame{
		Dmac:      net.HardwareAddr(b[:6:6]),
		Smac:      net.HardwareAddr(b[6:12:12]),
		EtherType: EtherType{b[12], b[13]},
		Payload:   b[14:],
	}
}
