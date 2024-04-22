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

const messageLen int = 1024 * 5

func main() {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(200, 0, 0, 1),
		Port: 40000,
	})
	if err != nil {
		utils.ErrorPrintf("connect to server err:", err)
		return
	}
	defer socket.Close()
	// generate random byte.
	bigMessage := make([]byte, messageLen)
	_, err = rand.Read(bigMessage)
	if err != nil {
		utils.ErrorPrintf("generate random byte err:%v\n", err)
	}

	go func() {
		i := 0
		for {
			data := bigMessage
			n, err := socket.Write(data)
			if err != nil {
				fmt.Printf("socket write err:%v\n", err)
			}
			utils.InfoPrintf("send %d byte.\n", n)
			utils.InfoPrintf("send data: %x\n", data)
			i += n
			if i == 5120 {
				break
			}
		}
	}()
	for {

	}

}
