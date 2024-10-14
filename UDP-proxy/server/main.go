package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type conf struct {
	LocalAddr  string `yaml:"localAddr"`
	RemoteAddr string `yaml:"remoteAddr"`
}

func main() {
	yamlFile, err := os.ReadFile("../../config.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}
	var c conf
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		fmt.Printf("yaml unmarshal err:%v\n", err)
	}
	fmt.Printf("config: %+v\n", c)
	// create UDP socket to proxy
	raddr, _ := net.ResolveUDPAddr("udp4", c.RemoteAddr)
	laddr, _ := net.ResolveUDPAddr("udp4", c.LocalAddr)

	socket, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Printf("UDP socket create err:%v\n", err)
	}
	data := make([]byte, 1500)
	n, _ := socket.Read(data)
	fmt.Println("receive tuple:")
	fmt.Println(string(data[:n]))
	// iperf server port is 8000
	// set the UDP socket
	ip := net.ParseIP("201.0.0.1")
	iperfSocket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ip,
		Port: 8000,
	})
	if err != nil {
		fmt.Println("create UDP socket to iperf server err")
	}
	// uplink
	go func() {
		data := make([]byte, 1500)
		for {
			n, err := socket.Read(data)
			if err != nil {
				fmt.Printf("UDP socket read err:%v\n", err)
			}
			iperfSocket.Write(data[:n])
		}
	}()
	// downlink
	i := 0
	go func() {
		data := make([]byte, 1500)
		for {
			n, err := iperfSocket.Read(data)
			if err != nil {
				fmt.Printf("iperfSocket read err:%v\n", err)
				break
			}
			_, err = socket.Write(data[:n])
			if err != nil {
				fmt.Printf("socket write err:%v\n", err)
			}
			i++
		}
		fmt.Printf("total receive %d packet.\n", i)
	}()

	for {

	}

}
