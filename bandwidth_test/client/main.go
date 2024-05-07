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

const messageLen int = 1024 * 1000

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
		j := 0
		for {
			// can't send more than 65535 bytes a time.
			j++
			data := bigMessage[i : i+65500]
			n, err := socket.Write(data)
			if err != nil {
				fmt.Printf("socket write err:%v\n", err)
				break
			}
			utils.InfoPrintf("the %d time, send %d byte.\n", j, n)
			i += n
			if i+65500 >= messageLen {
				break
			}
		}
		data := bigMessage[i:]
		n, err := socket.Write(data)
		if err != nil {
			fmt.Printf("socket write err:%v\n", err)
		}
		utils.InfoPrintf("send %d byte.\n", n)
		utils.InfoPrintf("total send %d byte. \n", len(bigMessage))
	}()
	for {

	}

}
