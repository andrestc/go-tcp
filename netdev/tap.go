package netdev

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/andrestc/go-tcp/arp"
)

const (
	SizeOfIfReq = 40
	IFNAMSIZ    = 16
	BUFFERSIZE  = 1522
)

var (
	tapFile    = "/dev/net/tap"
	deviceAddr = "10.0.0.5"
	devideCidr = "10.0.0.0/24"
	ARP        = EtherType{0x08, 0x06}
)

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

func (f *EthernetFrame) FromBytes(b []byte) *EthernetFrame {
	return &EthernetFrame{
		Dmac:      net.HardwareAddr(b[:6:6]),
		Smac:      net.HardwareAddr(b[6:12:12]),
		EtherType: EtherType{b[12], b[13]},
		Payload:   b[14:],
	}
}

type TAP struct {
	Addr string
	io.ReadWriteCloser
}

func (t *TAP) ReceiveLoop(ch chan<- []byte, done chan bool) {
	for {
		select {
		case <-done:
			fmt.Printf("Exiting receive loop\n")
			close(ch)
			return
		default:
			buffer := make([]byte, BUFFERSIZE)
			n, err := t.Read(buffer)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read dev file: %s", err)
				continue
			}
			if n > 0 {
				fmt.Printf("Read %d bytes from device\n", n)
				ch <- buffer[:n:n]
			}
		}
	}
}

func Handle(raw []byte) error {
	f := &EthernetFrame{}
	f.FromBytes(raw)
	switch f.EtherType {
	case ARP:
		return arp.Handle(f.Payload)
	default:
		fmt.Printf("Not implemented. Ignoring.\n")
	}
	return nil
}

type ifReq struct {
	Name  [IFNAMSIZ]byte
	Flags uint16
	pad   [SizeOfIfReq - IFNAMSIZ - 2]byte
}

func Init() (*TAP, error) {
	fmt.Println("Initilazing TAP device.")
	if _, err := os.Stat(tapFile); os.IsNotExist(err) {
		fmt.Println("Creating tap device file")
		errRun := runCmd("mknod", tapFile, "c", "10", "200")
		if errRun != nil {
			return nil, fmt.Errorf("failed to create tap file device: %s", errRun)
		}
	}
	tap, err := os.OpenFile(tapFile, syscall.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open tap dev file: %s", err)
	}
	ifReq := &ifReq{
		Flags: uint16(syscall.IFF_TAP) | uint16(syscall.IFF_NO_PI),
	}
	err = ioctl(uintptr(tap.Fd()), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(ifReq)))
	if err != nil {
		tap.Close()
		return nil, fmt.Errorf("failed to run ioctl: %s", err)
	}
	n := bytes.Index(ifReq.Name[:], []byte{0})
	err = configureDeviceInterface(string(ifReq.Name[:n]))
	if err != nil {
		tap.Close()
		return nil, fmt.Errorf("failed to configure dev interface: %s", err)
	}
	return &TAP{ReadWriteCloser: tap, Addr: deviceAddr}, nil
}

func configureDeviceInterface(dev string) error {
	fmt.Printf("Configuring %s dev interface\n", dev)
	if err := setIfaceUp(dev); err != nil {
		return fmt.Errorf("failed to set iface up: %s", err)
	}
	if err := setIfRoute(dev, devideCidr); err != nil {
		return fmt.Errorf("failed to set iface route: %s", err)
	}
	if err := setIfaceAddr(dev, deviceAddr); err != nil {
		return fmt.Errorf("failed to set iface addr: %s", err)
	}
	return nil
}

func setIfaceUp(dev string) error {
	return runCmd("ip", "link", "set", "dev", dev, "up")
}

func setIfaceAddr(dev, cidr string) error {
	return runCmd("ip", "address", "add", "dev", dev, "local", cidr)
}

func setIfRoute(dev, cidr string) error {
	return runCmd("ip", "route", "add", "dev", dev, cidr)
}

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}
