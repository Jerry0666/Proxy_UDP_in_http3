package main

import (
	"RFC9298proxy/utils"
	"net"
	"time"
)

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
	go func() {
		for {
			sendData := []byte("Hello server")
			_, err = socket.Write(sendData)
			if err != nil {
				utils.ErrorPrintf("socket write err:%v\n", err)
				continue
			}
			time.Sleep(time.Second)
		}
	}()

	receiveData := make([]byte, 1024)
	go func() {
		for {
			n, err := socket.Read(receiveData)
			if err != nil {
				utils.ErrorPrintf("socket read err:%v\n", err)
			}
			utils.InfoPrintf("got: %s\n", receiveData[:n])
		}
	}()

	for {

	}

}
