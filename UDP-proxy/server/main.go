package main

import (
	"fmt"
	"net"
)

func main() {
	// create UDP socket to proxy
	raddr, _ := net.ResolveUDPAddr("udp4", "172.16.0.3:9000")
	laddr, _ := net.ResolveUDPAddr("udp4", "192.168.5.1:7000")

	socket, _ := net.DialUDP("udp", laddr, raddr)
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
			n, _ := socket.Read(data)
			iperfSocket.Write(data[:n])
		}
	}()
	// downlink
	go func() {
		data := make([]byte, 1500)
		for {
			n, _ := iperfSocket.Read(data)
			socket.Write(data[:n])
		}

	}()
	for {

	}

}
