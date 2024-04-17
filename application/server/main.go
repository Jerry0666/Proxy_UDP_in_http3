package main

import (
	"fmt"
	"net"
)

func main() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 40000,
	})
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}
	defer listen.Close()
	data := make([]byte, 1024)
	for {
		n, addr, err := listen.ReadFromUDP(data)
		if err != nil {
			fmt.Println("read udp failed, err:", err)
			continue
		}
		fmt.Printf("got: %s\n", data[:n])
		listen.WriteToUDP([]byte("server got data"), addr)
	}
}
