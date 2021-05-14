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
	if proxyProtocol {
		remoteTmp, err := net.ResolveTCPAddr("tcp", conn.RemoteAddr().String())
		if err != nil {
			return
		}
		localTmp, err := net.ResolveTCPAddr("tcp", conn.LocalAddr().String())
		if err != nil {
			return
		}
		proxyProtocolText := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n", remoteTmp.IP.String(), localTmp.IP.String(), remoteTmp.Port, localTmp.Port)
		_, err = dialConn.Write([]byte(proxyProtocolText))
		if err != nil {
			return
		}
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

	if test {
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
	} else {
		_, _ = Transfer(right, left)
	}
}
