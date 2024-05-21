package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
)

func main() {
	socket, _ := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(100, 0, 0, 1),
		Port: 31000,
	})

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
	set := false
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			fmt.Println("tun read error")
		}
		if IsIPv4(buf[:n]) && IsUDP(buf[:n]) {
			if !set {
				src, dst := setUDPaddr(buf[:28])
				go downlink(socket, src, dst, ifce)
				set = true
			}
			CheckFragment(buf[:20])
			ip := ParseTargetIP(buf[:20])
			port := ParseTargetPort(buf[:28])
			if ip == "201.0.0.1" && port == "7000" {
				socket.Write(buf[28:n])
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

func ParseSourceIP(buf []byte) string {
	IPbyte := buf[12:16]
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

func ParseSourcePort(buf []byte) string {
	Portbyte := buf[20:22]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	return strconv.Itoa(port)
}

func setUDPaddr(buf []byte) (src, dst *net.UDPAddr) {
	targetIP := ParseTargetIP(buf)
	targetPort := ParseTargetPort(buf)
	targetPortInt, _ := strconv.Atoi(targetPort)
	sourceIP := ParseSourceIP(buf)
	sorcePort := ParseSourcePort(buf)
	sorcePortInt, _ := strconv.Atoi(sorcePort)

	src = &net.UDPAddr{
		IP:   net.ParseIP(sourceIP),
		Port: sorcePortInt,
	}
	dst = &net.UDPAddr{
		IP:   net.ParseIP(targetIP),
		Port: targetPortInt,
	}
	return src, dst
}

func buildUDPPacket(dst, src *net.UDPAddr, data []byte) ([]byte, error) {
	buffer := gopacket.NewSerializeBuffer()
	payload := gopacket.Payload(data)
	ip := &layers.IPv4{
		DstIP:    dst.IP,
		SrcIP:    src.IP,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(src.Port),
		DstPort: layers.UDPPort(dst.Port),
	}
	if err := udp.SetNetworkLayerForChecksum(ip); err != nil {
		return nil, fmt.Errorf("failed calc checksum: %s", err)
	}
	if err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}, ip, udp, payload); err != nil {
		return nil, fmt.Errorf("failed serialize packet: %s", err)
	}
	return buffer.Bytes(), nil
}

func downlink(socket *net.UDPConn, appClient, appServer *net.UDPAddr, ifce *water.Interface) {
	for {
		data := make([]byte, 1500)
		_, err := socket.Read(data)
		if err != nil {
			fmt.Println("UDP socket read err")
		}

		UDPpacket, err := buildUDPPacket(appClient, appServer, data)
		if err != nil {
			fmt.Println("buildUDPPacket err")
		} else {
			_, err := ifce.Write(UDPpacket)
			if err != nil {
				fmt.Println("ifce write err")
			}
		}
	}
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
