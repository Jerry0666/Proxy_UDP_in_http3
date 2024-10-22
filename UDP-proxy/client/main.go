package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"gopkg.in/yaml.v3"
)

type conf struct {
	LocalAddr  string `yaml:"localAddr"`
	RemoteAddr string `yaml:"remoteAddr"`
	// should write to CSV?
	WriteCSV bool `yaml:"writeCSV"`
}

var c conf

func main() {
	yamlFile, err := os.ReadFile("../../config.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err  %v\n", err)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		fmt.Printf("yaml unmarshal err:%v\n", err)
	}
	fmt.Printf("config: %+v\n", c)
	// create UDP socket to proxy
	laddr, _ := net.ResolveUDPAddr("udp4", c.LocalAddr)
	raddr, _ := net.ResolveUDPAddr("udp4", c.RemoteAddr)

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

	//uplink
	cycleTime := make([]int64, 5000)
	var LastTime, t1 time.Time
	LastTime = time.Now()
	i := 0
	go func() {
		time.Sleep(60 * time.Second)
		if c.WriteCSV {
			fmt.Println("Write to CSV...")
			file, err := os.OpenFile("../../uplinkCycle.csv", os.O_WRONLY, 0777)
			if err != nil {
				fmt.Printf("open csv file err:%v\n", err)
			}
			w := csv.NewWriter(file)
			w.Write([]string{"udp"})
			for i = 0; i < 5000; i++ {
				row := make([]string, 1)
				row[0] = strconv.Itoa(int(cycleTime[i]))
				w.Write(row)
			}
			w.Flush()
		}
	}()

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
					socket.Write(buf[28:n])
					continue
				}
				if ConnString == ConnTuple {
					socket.Write(buf[28:n])
					// record cycle time
					t1 = time.Now()
					if i < 5000 {
						// cycletime[0] should be remove.
						cycleTime[i] = int64(t1.Sub(LastTime))
					}
					LastTime = time.Now()
					i++
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
	cycleTime := make([]int64, 5000)
	var LastTime, t1 time.Time
	LastTime = time.Now()
	data := make([]byte, 1500)
	i := 0
	// go func() {
	// 	time.Sleep(60 * time.Second)
	// 	if c.WriteCSV {
	// 		fmt.Println("Write to CSV...")
	// 		file, err := os.OpenFile("../../cycle.csv", os.O_WRONLY, 0777)
	// 		if err != nil {
	// 			fmt.Printf("open csv file err:%v\n", err)
	// 		}
	// 		w := csv.NewWriter(file)
	// 		w.Write([]string{"cycle time"})
	// 		for i = 0; i < 5000; i++ {
	// 			row := make([]string, 1)
	// 			row[0] = strconv.Itoa(int(cycleTime[i]))
	// 			w.Write(row)
	// 		}
	// 		w.Flush()
	// 	}
	// }()
	for {
		n, _ := socket.Read(data)
		// record cycle time
		t1 = time.Now()
		if i < 5000 {
			// cycletime[0] should be remove.
			cycleTime[i] = int64(t1.Sub(LastTime))
		}
		LastTime = time.Now()
		UDPpacket, err := buildUDPPacket(appClient, appServer, data[:n])
		if err != nil {
			fmt.Printf("build UDP Packet err:%v\n", err)
			continue
		}
		_, err = ifce.Write(UDPpacket)
		if err != nil {
			fmt.Printf("ifce write err:%v\n", err)
		}
		i++
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
