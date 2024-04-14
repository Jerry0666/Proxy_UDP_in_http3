package main

import (
	"errors"
	"fmt"
	"net/http"

	"crypto/tls"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const proxyHost = "127.0.0.1"
const proxyPort = "30000"

func doProxyReq(client *http.Client, targetHost string, targetPort string) (int, error) {
	URL := "https://"
	URL += proxyHost
	URL += ":"
	URL += proxyPort
	URL += "/.well-known/masque/udp/"
	URL += targetHost
	URL += "/"
	URL += targetPort
	URL += "/"
	req, err := http.NewRequest(http.MethodConnect, URL, nil)
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return 0, err
	}
	roundTripper, ok := client.Transport.(*http3.RoundTripper)
	if !ok {
		err := errors.New("doProxyReq retrive roundTripper error.")
		return 0, err
	}
	roundTripper.AssignReqId(req)
	id, _ := http3.GetReqId(req)
	req.Proto = "connect-udp"
	go client.Do(req)
	return id, nil

}

func main() {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QuicConfig: &quic.Config{
			EnableDatagrams: true,
		},
		EnableDatagrams: true,
		ReqId:           0,
	}
	client := &http.Client{
		Transport: roundTripper,
	}
	defer roundTripper.Close()
	roundTripper.InitialMap()
	id, _ := doProxyReq(client, "8.8.8.8", "6666")
	str := roundTripper.GetReqStream(id)
	d, _ := str.Datagrammer()
	d.SendMessage([]byte("proxy test..."))

	for {

	}

}
