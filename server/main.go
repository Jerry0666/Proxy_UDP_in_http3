package main

import (
	"RFC9298proxy/proxy"
	"RFC9298proxy/utils"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

var clientChan chan *proxy.ProxyClient

func main() {
	clientChan = make(chan *proxy.ProxyClient, 10)
	go proxyHandler()
	datagramHandle := http.HandlerFunc(handle)
	server := http3.Server{
		Handler:   datagramHandle,
		Addr:      "0.0.0.0:30000",
		TLSConfig: http3.ConfigureTLSConfig(generateTLSConfig()), // use your tls.Config here
		QuicConfig: &quic.Config{
			KeepAlivePeriod: time.Minute * 5,
			EnableDatagrams: true,
		},
		EnableDatagrams: true,
		C1:              make(chan struct{}),
	}

	server.ListenAndServe()

}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("http handle function")
	fmt.Printf("URL PATH :%v\n", r.URL.Path)
	cl := new(proxy.ProxyClient)
	cl.SetUDPconn("201.0.0.1", "7000")
	s := r.Body.(http3.HTTPStreamer).HTTPStream()
	cl.Stream = s
	cl.Datagrammer, _ = s.Datagrammer()
	cl.Conn = cl.Datagrammer.GetQuicConn()
	clientChan <- cl
	path := r.URL.Path
	split := strings.Split(path, "/")
	utils.DebugPrintf("target host:%s\n", split[4])
	utils.DebugPrintf("target port:%s\n", split[5])
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()
}

func proxyHandler() {
	i := 0
	for {
		i++
		cl := <-clientChan
		fmt.Println("get proxy client")
		if i == 2 {
			go cl.DownlinkHandler()
			go cl.UplinkHandler()
		}
	}

}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	kl, err := os.OpenFile("tls_key.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
		KeyLogWriter: kl,
	}
}
