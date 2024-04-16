package proxy

import (
	"context"
	"fmt"

	"github.com/quic-go/quic-go/http3"
)

//add some structure to manage atream

type ProxyClient struct {
	Stream      http3.Stream
	Datagrammer http3.Datagrammer
}

func (c *ProxyClient) UplinkHandler() {
	for {
		data, err := c.Datagrammer.ReceiveMessage(context.Background())
		if err != nil {
			fmt.Printf("UplinkHandler err:%v\n", err)
		}
		fmt.Printf("data:%s\n", data)
	}
}
