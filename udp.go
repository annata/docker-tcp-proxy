package main

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

func udp(wg *sync.WaitGroup, port, start, length int) {
	defer wg.Done()
	listen, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		printError(err)
		return
	}
	defer listen.Close()
	for {
		data := make([]byte, 1518)
		n, addr, err := listen.ReadFromUDP(data)
		if err == nil {
			if n > 0 {
				go udpHandle(listen, data[0:n], addr, start, length)
			}
		} else {
			printError(err)
			break
		}
	}
}

func udpHandle(listen *net.UDPConn, data []byte, source *net.UDPAddr, start, length int) {
	var addr *net.UDPAddr
	if len(udpAddrList) == 0 {
		addr = udpAddr
	} else {
		addr = udpAddrList[start+rand.Intn(length)]
	}

	con, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer con.Close()
	con.SetDeadline(time.Now().Add(time.Second * 3))
	_, e := con.Write(data)
	if e != nil {
		printError(e)
		return
	}
	b := make([]byte, 1518)
	n, err := con.Read(b)
	if err == nil && n > 0 {
		listen.WriteToUDP(b[0:n], source)
	}
}
