package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ARPIPv4 = ARPProtoType(0x0800)
)

type ARPProtoType uint16

func (p ARPProtoType) String() string {
	switch p {
	case ARPIPv4:
		return "IPv4"
	default:
		return fmt.Sprintf("%d", p)
	}
}

type ARPPackage struct {
	HWType    uint16
	ProtoType ARPProtoType
	HWSize    uint8
	ProtoSize uint8
	OpCode    uint16
	Data      []byte
}

type ARPipv4 struct {
	Smac net.HardwareAddr
	Sip  net.IP
	Dmac net.HardwareAddr
	Dip  net.IP
}

func (p *ARPipv4) String() string {
	if p == nil {
		return "<?>"
	}
	return fmt.Sprintf("src %s %s dst %s %s", p.Smac, p.Sip, p.Dmac, p.Dip)
}

func (p *ARPPackage) FromBytes(b []byte) {
	p.HWType = binary.BigEndian.Uint16(b[:2:2])
	p.ProtoType = ARPProtoType(binary.BigEndian.Uint16(b[2:4:4]))
	p.HWSize = uint8(b[4])
	p.ProtoSize = uint8(b[5])
	p.OpCode = binary.BigEndian.Uint16(b[6:8:8])
	p.Data = b[8:]
}

func (p *ARPPackage) IPv4Data() *ARPipv4 {
	if p.ProtoType == ARPIPv4 {
		pSize := p.ProtoSize
		return &ARPipv4{
			Smac: net.HardwareAddr(p.Data[:6:6]),
			Sip:  net.IP(p.Data[6 : 6+pSize : 6+pSize]),
			Dmac: net.HardwareAddr(p.Data[10:16:16]),
			Dip:  net.IP(p.Data[16 : 16+pSize : 16+pSize]),
		}
	}
	return nil
}

func (p *ARPPackage) String() string {
	return fmt.Sprintf("[ARP]: layer %d (%d) proto %s (%d) op %d data [%s]",
		p.HWType, p.HWSize, p.ProtoType, p.ProtoSize, p.OpCode, p.IPv4Data(),
	)
}

func handleARP(f *EthernetFrame) error {
	pkg := &ARPPackage{}
	pkg.FromBytes(f.Payload)
	fmt.Println(pkg)
	return nil
}
