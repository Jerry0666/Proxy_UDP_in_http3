package main

import (
	"RFC9298proxy/utils"
	"fmt"
	"net"
)

func main() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 7000,
	})
	if err != nil {
		utils.ErrorPrintf("listen failed, err:", err)
		return
	}
	defer listen.Close()

	for {
		data := make([]byte, 1300)
		n, _, err := listen.ReadFromUDP(data)
		if err != nil {
			utils.ErrorPrintf("read udp failed, err:", err)
			continue
		}
		fmt.Printf("server read %d byte.\n", n)
		fmt.Printf("data:%x\n", data)
	}

}
