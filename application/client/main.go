package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	go second()
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(200, 0, 0, 1),
		Port: 40000,
	})
	if err != nil {
		fmt.Println("connect to server err:", err)
		return
	}
	defer socket.Close()
	for {
		sendData := []byte("Hello server")
		_, err = socket.Write(sendData)
		if err != nil {
			fmt.Println("send err:", err)
			continue
		}
		time.Sleep(time.Second)
	}
}

func second() {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(150, 0, 0, 2),
		Port: 40000,
	})
	if err != nil {
		fmt.Println("connect to server err:", err)
		return
	}
	defer socket.Close()
	for {
		sendData := []byte("Hi server, it is from second.")
		_, err = socket.Write(sendData)
		if err != nil {
			fmt.Println("send err:", err)
			continue
		}
		time.Sleep(time.Second)
	}
}
