package proxy

import (
	"RFC9298proxy/utils"
	"fmt"
	"net"
	"strconv"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

//add some structure to manage atream

type ProxyClient struct {
	Stream      http3.Stream
	Datagrammer http3.Datagrammer
	UDPsocket   *net.UDPConn
	Conn        quic.Connection
}

const HttpDataLen = 1310

func (c *ProxyClient) UplinkHandler() {
	fmt.Println("get Qconn")
	d := c.Datagrammer
	for {
		data, err := d.ReceiveMessage()
		if err != nil {
			utils.ErrorPrintf("UplinkHandler err:%v\n", err)
		}
		//forward data
		n, err := c.UDPsocket.Write(data)
		if err != nil {
			utils.ErrorPrintf("UDP write err:", err)
			continue
		}
		utils.DebugPrintf("proxy uplink write %d\n", n)
	}
}

func (c *ProxyClient) DownlinkHandler() {
	data := make([]byte, 1500)
	d := c.Datagrammer
	fmt.Println("get Qconn")
	for {
		n, err := c.UDPsocket.Read(data)
		if err != nil {
			utils.ErrorPrintf("UDPsocket read err:%v\n", err)
		}
		// utils.InfoPrintf("proxy downlink got: %x\n", data[:n])
		err = d.SendMessage(data[:n])
		if err != nil {
			fmt.Println("downlink SendMessage error")
			utils.ErrorPrintf("downlink handler err:%v\n", err)
		}
	}
}

func (c *ProxyClient) SetUDPconn(targetIP string, targetPort string) {
	ip := net.ParseIP(targetIP)
	port, err := strconv.Atoi(targetPort)
	if err != nil {
		utils.ErrorPrintf("SetUDPconn get prot err:%v\n", err)
	}
	//create udp socket
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		utils.ErrorPrintf("connect to server err:", err)
		return
	}
	c.UDPsocket = socket

}
