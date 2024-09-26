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
	// create UDP socket to proxy
	laddr, _ := net.ResolveUDPAddr("udp4", "172.16.0.3:9000")
	raddr, _ := net.ResolveUDPAddr("udp4", "192.168.5.1:7000")

	socket, _ := net.DialUDP("udp", laddr, raddr)

	tapconfig := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "tun0",
		},
	}
	ifce, err := water.New(tapconfig)
	if err != nil {
		fmt.Printf("create new tun interface err:%v\n", err)
	}
	setRoute()

	go func() {
		buf := make([]byte, 1500)
		Ready := false
		// only one connection
		ConnString := ""
		for {
			n, err := ifce.Read(buf)
			if err != nil {
				fmt.Printf("tun read err:%v\n", err)
			}
			if (IsIPv4(buf[:n]) && IsUDP(buf[:n])) || !Ready {
				sourceIP := ParseSourceIP(buf[:28])
				if sourceIP == "0.0.0.0" {
					continue
				}
				sourcePort := ParseSourcePort(buf[:28])
				targetIP := ParseTargetIP(buf[:28])
				targetPort := ParseTargetPort(buf[:28])
				ConnTuple := ""
				ConnTuple += sourceIP
				ConnTuple += ":"
				ConnTuple += sourcePort
				ConnTuple += "->"
				ConnTuple += targetIP
				ConnTuple += ":"
				ConnTuple += targetPort
				if !Ready {
					fmt.Printf("ConnTuple:%s\n", ConnTuple)
					socket.Write([]byte(ConnTuple))
					ConnString = ConnTuple
					src, dst := setUDPaddr(buf[:28])
					go downlink(socket, src, dst, ifce)
					Ready = true
				}
				if ConnString == ConnTuple {
					socket.Write(buf[28:n])
				}

			}
		}
	}()

	for {

	}

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

func setRoute() {
	cmd := exec.Command("ip", "addr", "add", "12.0.0.1/24", "dev", "tun0")
	execCommand(cmd)
	cmd = exec.Command("ip", "link", "set", "tun0", "up")
	execCommand(cmd)
	cmd = exec.Command("ip", "r", "add", "201.0.0.1", "dev", "tun0")
	execCommand(cmd)
}

func execCommand(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		fmt.Println("exec command err")
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

func downlink(socket *net.UDPConn, appClient, appServer *net.UDPAddr, ifce *water.Interface) {
	data := make([]byte, 1500)
	for {
		n, _ := socket.Read(data)
		UDPpacket, err := buildUDPPacket(appClient, appServer, data[:n])
		if err != nil {
			fmt.Printf("build UDP Packet err:%v\n", err)
			continue
		}
		_, err = ifce.Write(UDPpacket)
		if err != nil {
			fmt.Printf("ifce write err:%v\n", err)
		}
	}
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
		return nil, fmt.Errorf("Failed calc checksum: %s", err)
	}
	if err := gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}, ip, udp, payload); err != nil {
		return nil, fmt.Errorf("Failed serialize packet: %s", err)
	}
	return buffer.Bytes(), nil
}
