package main

import (
	"errors"
	"fmt"
	"net/http"

	"crypto/tls"
	"os/exec"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

const proxyHost = "100.0.0.1"
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

func waterRead(ifce *water.Interface) {
	for {
		var buf ethernet.Frame
		buf.Resize(1500)
		n, err := ifce.Read(buf)
		if err != nil {
			fmt.Println("tun read err")
		}
		fmt.Printf("packet:%x\n", buf[:n])
		fmt.Printf("len:%d\n", n)
	}

}

func execCommand(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		fmt.Println("exec command error!")
	}
}

func setRoute() {
	cmd := exec.Command("ip", "addr", "add", "12.0.0.1/24", "dev", "tun0")
	execCommand(cmd)
	cmd = exec.Command("ip", "link", "set", "tun0", "up")
	execCommand(cmd)
	cmd = exec.Command("ip", "r", "replace", "default", "dev", "tun0")
	execCommand(cmd)
}

func main() {
	tapconfig := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "tun0",
		},
	}
	ifce, err := water.New(tapconfig)
	if err != nil {
		fmt.Println("create new tun interface error.")
	}
	go waterRead(ifce)
	setRoute()

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
