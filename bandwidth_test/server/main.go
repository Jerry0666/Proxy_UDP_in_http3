package main

import (
	"RFC9298proxy/utils"
	"net"
)

func main() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 40000,
	})
	if err != nil {
		utils.ErrorPrintf("listen failed, err:", err)
		return
	}
	defer listen.Close()
	data := make([]byte, 70000)
	bigMessage := make([]byte, 0)

	i := 0
	for {
		i++
		n, _, err := listen.ReadFromUDP(data)
		if err != nil {
			utils.ErrorPrintf("read udp failed, err:", err)
			continue
		}
		bigMessage = append(bigMessage, data[:n]...)
		utils.InfoPrintf("i=%d read %d byte, message len:%d\n", i, n, len(bigMessage))
	}

}
