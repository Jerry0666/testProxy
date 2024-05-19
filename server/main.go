package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	err := Server()
	if err != nil {
		fmt.Println("server err")
		fmt.Println(err)
	}

}

func Server() error {
	const addr = "100.0.0.1:30000"
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	listener, err := quic.ListenAddr(addr, generateTLSConfig(), &quic.Config{
		KeepAlivePeriod: time.Minute * 5,
		EnableDatagrams: true,
	})
	if err != nil {
		return err
	}

	conn, err := listener.Accept(ctx)
	if err != nil {
		return err
	}

	ip := net.ParseIP("201.0.0.1")
	port, _ := strconv.Atoi("7000")
	//create udp socket
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		fmt.Printf("connect to server err: %v \n", err)
	}
	for {

		buf, err := conn.ReceiveDatagram(context.Background())
		if err != nil {
			fmt.Println(err)
			break
		}
		_, err = socket.Write(buf)
		if err != nil {
			fmt.Printf("udp write err. %v\n", err)
		}

	}
	return err

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
