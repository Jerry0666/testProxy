package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	err := Server()
	if err != nil {
		fmt.Println("server err")
		fmt.Println(err)
	}

}

func Server() error {
	listen, _ := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 31000,
	})

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

	data := make([]byte, 1472)
	for {

		n, _, _ := listen.ReadFromUDP(data)
		_, err = socket.Write(data[:n])
		if err != nil {
			fmt.Printf("udp write err. %v\n", err)
		}

	}

}
