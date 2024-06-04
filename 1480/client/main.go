package main

import (
	"RFC9298proxy/utils"
	"crypto/rand"
	"fmt"
	"net"
)

// Test transmitting a large series of messages,
// create big message randomly, write it into file
// and send it to the server.

const messageLen int = 1300

func main() {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(201, 0, 0, 1),
		Port: 40000,
	})
	if err != nil {
		utils.ErrorPrintf("connect to server err:", err)
		return
	}
	defer socket.Close()

	go func() {
		for i := 0; i < 5; i++ {
			// generate random byte.
			data := make([]byte, messageLen)
			_, err = rand.Read(data)
			if err != nil {
				fmt.Printf("generate random byte err:%v\n", err)
			}
			n, err := socket.Write(data)
			if err != nil {
				fmt.Printf("socket write err:%v\n", err)
			}
			fmt.Printf("send %d byte.\n", n)
			fmt.Printf("data: %x\n", data)

		}
	}()
	for {

	}

}
