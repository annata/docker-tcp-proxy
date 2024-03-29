package main

import (
	"flag"
	"os"
	"strconv"
	"sync"
)

var (
	port          int
	portList      = make([]int, 0, 100)
	tcpAddr       string
	udpAddr       string
	tcpAddrList   = make([]string, 0, 100)
	udpAddrList   = make([]string, 0, 100)
	mode          int
	test          = false
	proxyProtocol = false
	proxyAddr     = ""
	proxyUser     = ""
)

func main() {
	if !parse() {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)

	if len(portList) == 0 {
		if mode == 0 || mode == 1 {
			go tcp(wg, port, 0, len(tcpAddrList))
		}
		if mode == 2 || mode == 1 {
			go udp(wg, port, 0, len(tcpAddrList))
		}
	} else {
		tcpAddrListLen := len(tcpAddrList)
		portListLen := len(portList)
		aa := tcpAddrListLen / portListLen
		bb := tcpAddrListLen % portListLen
		start := 0

		for i := 0; i < portListLen; i++ {
			var num int
			if i < bb {
				num = aa + 1
			} else {
				num = aa
			}
			if mode == 0 || mode == 1 {
				go tcp(wg, portList[i], start, num)
			}
			if mode == 2 || mode == 1 {
				go udp(wg, portList[i], start, num)
			}
			start += num
		}
	}
	wg.Wait()
}

func parse() bool {
	domain := ""
	proxyDomain := ""
	flag.IntVar(&port, "p", 0, "监听端口")
	flag.StringVar(&domain, "d", "", "访问的域名端口")
	flag.StringVar(&proxyDomain, "pd", "", "代理url")
	flag.StringVar(&proxyUser, "pu", "", "代理用户名密码")
	flag.IntVar(&mode, "m", 0, "转发模式,0为tcp,1为tcp+udp,2为udp.默认为0")
	flag.BoolVar(&test, "t", false, "测试模式")
	flag.BoolVar(&proxyProtocol, "proxy", false, "proxy_protocol模式")
	flag.Parse()
	modeStr := os.Getenv("MODE")
	if modeStr != "" {
		modeInt, err := strconv.Atoi(modeStr)
		if err != nil {
			printError(err)
			return false
		}
		mode = modeInt
	}
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
	proxyDomainStr := os.Getenv("PROXY_DOMAIN")
	if proxyDomainStr != "" {
		proxyDomain = proxyDomainStr
	}
	proxyUserStr := os.Getenv("PROXY_USER")
	if proxyUserStr != "" {
		proxyUser = proxyUserStr
	}
	if domain != "" {
		tcpAddr = domain
		udpAddr = domain
	}
	if proxyDomain != "" {
		proxyAddr = proxyDomain
	}

	for i := 0; true; i++ {
		d := os.Getenv("DOMAIN_" + strconv.Itoa(i))
		if d == "" {
			break
		} else {
			tcpAddrList = append(tcpAddrList, d)
			udpAddrList = append(udpAddrList, d)
		}
	}

	for i := 0; true; i++ {
		p := os.Getenv("PORT_" + strconv.Itoa(i))
		if p == "" {
			break
		} else {
			tmpP, e := strconv.Atoi(p)
			if e != nil {
				return false
			}
			if tmpP <= 0 || tmpP >= 65536 {
				return false
			}
			portList = append(portList, tmpP)
		}
	}

	testStr := os.Getenv("TEST")
	if testStr != "" {
		tests, e := strconv.ParseBool(testStr)
		if e != nil {
			return false
		}
		test = tests
	}

	proxyProtocolStr := os.Getenv("PROXY_PROTOCOL")
	if proxyProtocolStr != "" {
		proxyProtocols, e := strconv.ParseBool(proxyProtocolStr)
		if e != nil {
			return false
		}
		proxyProtocol = proxyProtocols
	}

	if (len(portList) == 0 && (port <= 0 || port >= 65536)) || (tcpAddr == "" && len(tcpAddrList) == 0) ||
		(udpAddr == "" && len(udpAddrList) == 0) || mode < 0 || mode > 2 {
		flag.Usage()
		return false
	}

	return true
}

func printError(err error) {
	//fmt.Println(err.Error())
}
