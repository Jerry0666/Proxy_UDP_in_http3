package proxy

import (
	"RFC9298proxy/utils"
	"context"
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

const HttpDataLen = 1310

func (c *ProxyClient) UplinkHandler() {
	bigData := make([]byte, 0)
	for {
		data, err := c.Datagrammer.HardcodedRead(context.Background())
		if err != nil {
			utils.ErrorPrintf("UplinkHandler err:%v\n", err)
		}
		dataLen := len(data)
		if dataLen == HttpDataLen+1 && data[HttpDataLen] == 0xff {
			bigData = append(bigData, data[0:HttpDataLen]...)
			continue
		} else {
			if len(bigData) != 0 {
				bigData = append(bigData, data...)
				_, err := c.UDPsocket.Write(bigData)
				if err != nil {
					utils.ErrorPrintf("UDP write err:", err)
					continue
				}
				bigData = make([]byte, 0)
				continue
			}
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
	data := make([]byte, 1024)
	for {
		n, err := c.UDPsocket.Read(data)
		if err != nil {
			utils.ErrorPrintf("UDPsocket read err:%v\n", err)
		}
		// utils.InfoPrintf("proxy downlink got: %x\n", data[:n])
		err = c.Datagrammer.SendMessage(data[:n])
		if err != nil {
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
