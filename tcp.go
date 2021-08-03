package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
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
	addr := ""
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
	realAddr := ""
	if proxyAddr == "" {
		realAddr = addr
	} else {
		realAddr = proxyAddr
	}

	tcpAddrTemp, err := net.ResolveTCPAddr("tcp", realAddr)
	if err != nil {
		return
	}
	dialConn, err := net.DialTCP("tcp", nil, tcpAddrTemp)
	if err != nil {
		printError(err)
		return
	}
	defer dialConn.Close()
	dialConnReader := bufio.NewReader(dialConn)
	if proxyAddr != "" {
		proxyText := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nProxy-Authorization: Basic %s\r\nProxy-Connection: Keep-Alive\r\n\r\n", addr, base64.StdEncoding.EncodeToString([]byte(proxyUser)))
		_, err = dialConn.Write([]byte(proxyText))
		if err != nil {
			return
		}
		success := false
		for {
			command, _, err := dialConnReader.ReadLine()
			if err != nil {
				break
			}
			if string(command) == "" {
				success = true
				break
			}
		}
		if !success {
			return
		}
	}
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
	go trans(wg, conn, dialConn, conn, dialConn)
	go trans(wg, dialConn, conn, dialConnReader, conn)
	wg.Wait()
}

func trans(wg *sync.WaitGroup, left, right *net.TCPConn, src io.Reader, dst io.Writer) {
	defer wg.Done()
	defer left.CloseRead()
	defer right.CloseWrite()

	if test {
		data := make([]byte, 1600)
		for {
			n, err := src.Read(data)
			if err != nil {
				printError(err)
				return
			}
			_, e := dst.Write(data[0:n])
			if e != nil {
				printError(e)
				return
			}
			if test {
				fmt.Println(hex.EncodeToString(data[0:n]))
			}
		}
	} else {
		_, _ = io.Copy(dst, src)
	}
}
