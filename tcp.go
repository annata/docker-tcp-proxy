package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"sync"
)

func tcp(wg *sync.WaitGroup, port, start, length int) {
	defer wg.Done()

	listen, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		printError(err)
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			printError(err)
			break
		}
		go handle(conn, start, length)
	}
}

func handle(conn *net.TCPConn, start, length int) {
	defer conn.Close()
	var addr *net.TCPAddr
	tcpAddrListLen := len(tcpAddrList)
	if tcpAddrListLen == 0 {
		addr = tcpAddr
	} else if length == 0 {
		return
	} else if length == 1 {
		addr = tcpAddrList[start]
	} else {
		addr = tcpAddrList[start+rand.Intn(length)]
	}
	dialConn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		printError(err)
		return
	}
	defer dialConn.Close()
	err = conn.SetNoDelay(true)
	if err != nil {
		printError(err)
		return
	}
	err = dialConn.SetNoDelay(true)
	if err != nil {
		printError(err)
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go trans(wg, conn, dialConn)
	go trans(wg, dialConn, conn)
	wg.Wait()
}

func trans(wg *sync.WaitGroup, left, right *net.TCPConn) {
	defer wg.Done()
	defer left.CloseRead()
	defer right.CloseWrite()
	data := make([]byte, 1600)
	for {
		n, err := left.Read(data)
		if err != nil {
			printError(err)
			return
		}
		_, e := right.Write(data[0:n])
		if e != nil {
			printError(e)
			return
		}
		if test {
			fmt.Println(hex.EncodeToString(data[0:n]))
		}

	}
}
