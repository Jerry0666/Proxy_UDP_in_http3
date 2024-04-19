package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/quic-go/quic-go/http3"
)

//add some structure to manage atream

type ProxyClient struct {
	Stream      http3.Stream
	Datagrammer http3.Datagrammer
	UDPsocket   *net.UDPConn
}

func (c *ProxyClient) UplinkHandler() {
	for {
		data, err := c.Datagrammer.ReceiveMessage(context.Background())
		if err != nil {
			fmt.Printf("UplinkHandler err:%v\n", err)
		}
		fmt.Printf("proxy uplink got: %s\n", data)
		//forward data
		_, err = c.UDPsocket.Write(data)
		if err != nil {
			fmt.Println("send err:", err)
			continue
		}
	}
}

func (c *ProxyClient) DownlinkHandler() {
	data := make([]byte, 1024)
	for {
		n, err := c.UDPsocket.Read(data)
		if err != nil {
			fmt.Printf("UDPsocket read err:%v\n", err)
		}
		fmt.Printf("proxy downlink got: %s\n", data[:n])
		err = c.Datagrammer.SendMessage(data[:n])
		if err != nil {
			fmt.Printf("downlink handler err:%v\n", err)
		}
	}
}

func (c *ProxyClient) SetUDPconn(targetIP string, targetPort string) {
	ip := net.ParseIP(targetIP)
	port, err := strconv.Atoi(targetPort)
	if err != nil {
		fmt.Printf("SetUDPconn get prot err:%v\n", err)
	}
	//create udp socket
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		fmt.Println("connect to server err:", err)
		return
	}
	c.UDPsocket = socket

}
