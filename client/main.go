package main

import (
	"RFC9298proxy/utils"
	"errors"
	"fmt"
	"net/http"
	"time"

	"crypto/tls"
	"net"
	"os/exec"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/songgao/water"
)

const proxyHost = "100.0.0.1"
const proxyPort = "30000"
const HttpDataLen = 1310
const TestIP = "201.0.0.1"
const TestPort = "7000"

func main() {
	udpaddr, _ := net.ResolveUDPAddr("udp4", "10.0.0.1:9000")
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-echo-example"},
		},
		QuicConfig: &quic.Config{
			KeepAlivePeriod: time.Minute * 5,
			EnableDatagrams: true,
		},
		EnableDatagrams: true,
		ReqId:           0,
		DefaultAddr:     udpaddr,
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
		utils.ErrorPrintf("create new tun interface err:%v\n", err)
	}
	setRoute()
	ProxyManager := make(map[string]http3.Datagrammer)
	var src, dst *net.UDPAddr
	doProxyReq(client, TestIP, "8000")
	// get quic connection transport
	tr := roundTripper.GetTransport()

	var Qconn quic.MPConnection = roundTripper.TempConn

	if Qconn == nil {
		fmt.Println("Qconn is nil")
	} else {
		fmt.Println("get the Qconn")
	}

	Qconn.SetTransport(tr)
	OriginalPath := Qconn.GetPath()

	fmt.Printf("[proxy]remote addr:%s\n", Qconn.RemoteAddr().String())
	path := quic.NewPath(tr, Qconn.RemoteAddr(), true)
	path.SetIP("11.0.0.1", 7000)
	path.Status = quic.PathStatusProbing

	// Qconn.ProbePath(tr)
	Qconn.SendPathChallenge(path)
	Qconn.RecordPath(path)

	// Add the control plane server
	controlServer, err := net.ListenPacket("udp", "127.0.0.1:8964")
	if err != nil {
		fmt.Println("open control socket error!")
	}

	go readCommand(controlServer, Qconn)

	//uplink
	go func() {

		i := 0
		buf := make([]byte, 1500)
		for {
			i++
			if i == 4000 {
				fmt.Println("i >= 4000, do Migration.")
				Qconn.Migration(path)
			}
			if i == 8000 {
				fmt.Println("i == 8000, migration back")
				Qconn.Migration(OriginalPath)
			}
			if i == 12000 {
				fmt.Println("i == 12000, migration")
				Qconn.Migration(path)
			}
			n, err := ifce.Read(buf)
			if err != nil {
				utils.ErrorPrintf("tun read err:%v\n", err)
			}
			utils.DebugPrintf("--------------------uplink read %d byte.--------------------\n", n)
			if IsIPv4(buf[:n]) && IsUDP(buf[:n]) {
				sourceIP := ParseSourceIP(buf[:28])
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
				d, ok := ProxyManager[ConnTuple]
				if !ok {
					fmt.Println("datagram not set yet.")
					fmt.Printf("connection four tuple is: %s \n", ConnTuple)
					rsp, _ := doProxyReq(client, targetIP, targetPort)
					if rsp == nil {
						fmt.Println("rsp is nil!")
						break
					}
					s := rsp.Body.(http3.HTTPStreamer).HTTPStream()
					fmt.Println("get the http stream")
					d, _ = s.Datagrammer()
					ProxyManager[ConnTuple] = d
					fmt.Println("get the datagram")
					src, dst = setUDPaddr(buf[:28])
					// create the downlink go routine
					go downlink(d, src, dst, ifce)
				}
				err := d.SendMessage(buf[28:n])
				if err != nil {
					fmt.Printf("send Message err:%v\n", err)
				}
			}
		}
	}()
	for {

	}

}

// read control command
func readCommand(c net.PacketConn, QConn quic.MPConnection) {
	fmt.Println("read command start...")
	buf := make([]byte, 1024)
	for {
		n, addr, err := c.ReadFrom(buf)
		if err != nil {
			fmt.Println("[control] read error")
		}
		convertString := string(buf[:n])
		prefix := convertString[:6]
		if prefix == "[task]" {
			fmt.Println("find a task")
			suffix := convertString[6:n]
			switch suffix {
			case " check all path status":
				fmt.Println("check the path status")
				status := QConn.CheckStatus()
				fmt.Printf("%s", status)
				c.WriteTo([]byte(status), addr)
			}
		}

	}

}

func doProxyReq(client *http.Client, targetHost string, targetPort string) (*http.Response, error) {
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
		utils.ErrorPrintf("http do request err:%v\n", err)
		return nil, err
	}
	roundTripper, ok := client.Transport.(*http3.RoundTripper)
	if !ok {
		err := errors.New("doProxyReq retrive roundTripper error.")
		return nil, err
	}
	roundTripper.AssignReqId(req)
	req.Proto = "connect-udp"
	return client.Do(req)
}

func execCommand(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		utils.ErrorPrintf("exec command error!")
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
	utils.DebugLog("target ip:%s\n", IPstring)
	return IPstring
}

func ParseTargetPort(buf []byte) string {
	Portbyte := buf[22:24]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	utils.DebugLog("target port:%d\n", port)
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
	utils.DebugLog("source ip:%s\n", IPstring)
	return IPstring
}

func ParseSourcePort(buf []byte) string {
	Portbyte := buf[20:22]
	port := int(Portbyte[0]) * 256
	port += int(Portbyte[1])
	utils.DebugLog("source port:%d\n", port)
	return strconv.Itoa(port)
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

func downlink(d http3.Datagrammer, appClient, appServer *net.UDPAddr, ifce *water.Interface) {

	for {
		data, err := d.ReceiveMessage()
		if err != nil {
			fmt.Println("downlink datagram receive message err")
			utils.ErrorPrintf("downlink datagram receive message err:%v\n", err)
		}
		utils.InfoLog("proxy client downlink got: %s\n", data)
		UDPpacket, err := buildUDPPacket(appClient, appServer, data)
		if err != nil {
			utils.ErrorPrintf("build UDP packet err: %v\n", err)
		} else {
			_, err := ifce.Write(UDPpacket)
			if err != nil {
				utils.ErrorPrintf("ifce Write err:%v\n", err)
			}
		}
	}
}
