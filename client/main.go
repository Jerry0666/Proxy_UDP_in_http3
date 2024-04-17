package main

import (
	"errors"
	"fmt"
	"net/http"

	"crypto/tls"
	"net"
	"os/exec"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

const proxyHost = "100.0.0.1"
const proxyPort = "30000"

func doProxyReq(client *http.Client, targetHost string, targetPort string) (int, error) {
	URL := "https://"
	URL += proxyHost
	URL += ":"
	URL += proxyPort
	URL += "/.well-known/masque/udp/"
	URL += targetHost
	URL += "/"
	URL += targetPort
	URL += "/"
	req, err := http.NewRequest(http.MethodConnect, URL, nil)
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return 0, err
	}
	roundTripper, ok := client.Transport.(*http3.RoundTripper)
	if !ok {
		err := errors.New("doProxyReq retrive roundTripper error.")
		return 0, err
	}
	roundTripper.AssignReqId(req)
	id, _ := http3.GetReqId(req)
	req.Proto = "connect-udp"
	go client.Do(req)
	return id, nil

}

func waterRead(ifce *water.Interface) {
	for {
		var buf ethernet.Frame
		buf.Resize(1500)
		n, err := ifce.Read(buf)
		if err != nil {
			fmt.Println("tun read err")
		}
		fmt.Printf("packet:%x\n", buf[:n])
		fmt.Printf("len:%d\n", n)
	}

}

func execCommand(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		fmt.Println("exec command error!")
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
	fmt.Printf("target ip:%s\n", IPstring)
	return IPstring
}

func ParseTargetPort(buf []byte) string {
	Portbyte := buf[22:24]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	fmt.Printf("target port:%d\n", port)
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
	fmt.Printf("source ip:%s\n", IPstring)
	return IPstring
}

func ParseSourcePort(buf []byte) string {
	Portbyte := buf[20:22]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	fmt.Printf("source port:%d\n", port)
	return strconv.Itoa(port)
}

func buildUDPPacket(dst, src *net.UDPAddr) ([]byte, error) {
	buffer := gopacket.NewSerializeBuffer()
	payload := gopacket.Payload("build udp raw packet testing...")
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

func main() {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QuicConfig: &quic.Config{
			EnableDatagrams: true,
		},
		EnableDatagrams: true,
		ReqId:           0,
	}
	client := &http.Client{
		Transport: roundTripper,
	}
	defer roundTripper.Close()
	roundTripper.InitialMap()
	tapconfig := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "tun0",
		},
	}
	ifce, err := water.New(tapconfig)
	if err != nil {
		fmt.Println("create new tun interface error.")
	}
	setRoute()
	ProxyManager := make(map[string]http3.Datagrammer)
	//test, temporary variable
	var src, dst *net.UDPAddr
	waitchan := make(chan struct{})
	//uplink
	go func() {
		for {
			var buf ethernet.Frame
			buf.Resize(1024)
			n, err := ifce.Read(buf)
			if err != nil {
				fmt.Println("tun read err")
			}
			if IsIPv4(buf[:n]) && IsUDP(buf[:n]) {
				targetIP := ParseTargetIP(buf[:n])
				targetPort := ParseTargetPort(buf[:n])
				targetAddr := targetIP + ":" + targetPort
				d, ok := ProxyManager[targetAddr]
				if !ok {
					//do a proxy request and get the datagrammer.
					id, _ := doProxyReq(client, targetIP, targetPort)
					str := roundTripper.GetReqStream(id)
					d, _ := str.Datagrammer()
					ProxyManager[targetAddr] = d
					//set the udp addr
					src, dst = setUDPaddr(buf[:n])
					waitchan <- struct{}{}
				} else {
					d.SendMessage(buf[28:n])
				}
			}
			fmt.Printf("packet:%x\n", buf[:n])
		}
	}()
	//downlink
	go func() {
		<-waitchan
		//downlink, switch src snd dst
		data, err := buildUDPPacket(src, dst)
		if err != nil {
			fmt.Printf("buildUDPPacket err:%v\n", err)
		} else {
			_, err := ifce.Write(data)
			if err != nil {
				fmt.Printf("ifce Write err:%v\n", err)
			}
		}

	}()

	for {

	}

}
