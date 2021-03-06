package arp

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ARPIPv4 = ARPProtoType(0x0800)

	ARPEthernet = ARPHWType(0x0001)

	ARPRequest = uint16(1)
	ARPReply   = uint16(2)
)

type ARPHWType uint16

func (h ARPHWType) String() string {
	switch h {
	case ARPEthernet:
		return "Ethernet"
	default:
		return fmt.Sprintf("%d", h)
	}
}

type ARPProtoType uint16

func (p ARPProtoType) String() string {
	switch p {
	case ARPIPv4:
		return "IPv4"
	default:
		return fmt.Sprintf("%d", p)
	}
}

type ARPFrame struct {
	HWType    ARPHWType
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

func fromBytes(b []byte) *ARPFrame {
	return &ARPFrame{
		HWType:    ARPHWType(binary.BigEndian.Uint16(b[:2:2])),
		ProtoType: ARPProtoType(binary.BigEndian.Uint16(b[2:4:4])),
		HWSize:    uint8(b[4]),
		ProtoSize: uint8(b[5]),
		OpCode:    binary.BigEndian.Uint16(b[6:8:8]),
		Data:      b[8:],
	}
}

func (p *ARPFrame) IPv4Data() *ARPipv4 {
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

func (p *ARPFrame) String() string {
	return fmt.Sprintf("[ARP]: hw %s (%d) proto %s (%d) op %d data [%s]",
		p.HWType, p.HWSize, p.ProtoType, p.ProtoSize, p.OpCode, p.IPv4Data(),
	)
}

func Handle(f []byte) error {
	frame := fromBytes(f)
	fmt.Println(frame)
	if frame.HWType != ARPEthernet {
		return fmt.Errorf("unsuported HW type: %s", frame.HWType)
	}
	if frame.ProtoType != ARPIPv4 {
		return fmt.Errorf("unsuported protocol: %s", frame.ProtoType)
	}
	if frame.OpCode != ARPRequest {
		return fmt.Errorf("unsuported ARP operation: %d", frame.OpCode)
	}
	return replyARP(frame)
}

func replyARP(p *ARPFrame) error {
	return nil
}
