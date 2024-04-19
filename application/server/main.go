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
	data := make([]byte, 1024)
	for {
		n, addr, err := listen.ReadFromUDP(data)
		if err != nil {
			utils.ErrorPrintf("read udp failed, err:", err)
			continue
		}
		utils.InfoPrintf("got: %s\n", data[:n])
		listen.WriteToUDP([]byte("server got data"), addr)
	}
}
