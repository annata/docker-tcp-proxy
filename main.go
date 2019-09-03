package main

import (
	"net"
	"sync"
	"flag"
	"os"
	"strconv"
)

var (
	domain  string
	port    int
	tcpAddr *net.TCPAddr
)

func main() {
	//rand.Seed(time.Now().UnixNano())
	if !parse() {
		return
	}

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
		go handle(conn)
	}
}

func parse() bool {
	flag.IntVar(&port, "p", 0, "监听端口")
	flag.StringVar(&domain, "d", "", "访问的域名端口")
	flag.Parse()
	portStr := os.Getenv("PORT")
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			printError(err)
			return false
		}
		port = portInt
	}
	domainStr := os.Getenv("DOMAIN")
	if domainStr != "" {
		domain = domainStr
	}

	if port <= 0 || port >= 65536 || domain == "" {
		flag.Usage()
		return false
	}

	addr, err := net.ResolveTCPAddr("tcp", domain)
	if err != nil {
		printError(err)
		return false
	}
	tcpAddr = addr
	return true
}

func handle(conn *net.TCPConn) {
	defer conn.Close()
	dialConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		printError(err)
		return
	}
	defer dialConn.Close()
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
	data := make([]byte, 1518)
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
	}
}

func printError(err error) {
	//fmt.Println(err.Error())
}
