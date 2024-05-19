package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/songgao/water"

	"github.com/quic-go/quic-go"
)

const addr = "100.0.0.1:30000"

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, &quic.Config{
		KeepAlivePeriod: time.Minute * 5,
		EnableDatagrams: true,
	})
	if err != nil {
		fmt.Println("conn err")
	}
	tunconfig := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "tun0",
		},
	}
	ifce, err := water.New(tunconfig)
	if err != nil {
		fmt.Println("create tun error")
	}
	setRoute()
	buf := make([]byte, 1500)
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			fmt.Println("tun read error")
		}
		if IsIPv4(buf[:n]) && IsUDP(buf[:n]) {
			CheckFragment(buf[:20])
			ip := ParseTargetIP(buf[:20])
			port := ParseTargetPort(buf[:28])
			if ip == "201.0.0.1" && port == "7000" {
				conn.SendDatagram(buf[28:n])
			}
		}
	}
}

func CheckFragment(buf []byte) bool {
	flag := (buf[6] & 0x20)
	if flag == 0x20 {
		fmt.Println("More Fragment bit is set!")
		return false
	}
	return true
}

func IsIPv4(buf []byte) bool {
	if buf[0] == 0x45 {
		return true
	} else {
		return false
	}
}

func IsUDP(buf []byte) bool {
	if buf[9] == 0x11 {
		return true
	} else {
		return false
	}
}

func ParseTargetIP(buf []byte) string {
	IPbyte := buf[16:20]
	IPstring := ""
	for i := 0; i < 3; i++ {
		x := int(IPbyte[i])
		IPstring += strconv.Itoa(x)
		IPstring += "."
	}
	x := int(IPbyte[3])
	IPstring += strconv.Itoa(x)
	return IPstring
}

func ParseTargetPort(buf []byte) string {
	Portbyte := buf[22:24]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	return strconv.Itoa(port)
}

func execCommand(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd exec error")
	}
}

func setRoute() {
	cmd := exec.Command("ip", "addr", "add", "12.0.0.1/24", "dev", "tun0")
	execCommand(cmd)
	cmd = exec.Command("ip", "link", "set", "tun0", "up")
	execCommand(cmd)
	cmd = exec.Command("ip", "r", "replace", "default", "dev", "tun0")
	execCommand(cmd)
}
