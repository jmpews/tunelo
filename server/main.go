package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"tunelo/pkg/logger/plain"
)

func main() {
	var serverIP string
	var serverPort string
	var vpnListenPort string
	var protocol string

	flag.StringVar(&serverIP, "server_ip", "127.0.0.1", "Proxy server IP address.")
	flag.StringVar(&serverPort, "server_port", "11821", "Proxy server port number.")
	flag.StringVar(&vpnListenPort, "vpn_listen_port", "11820", "VPN listen port number.")
	flag.StringVar(&protocol, "p", "ws", "Tunnel transport protocol. Options: ws, utls, and tcp.")
	flag.Parse()

	logger := plain.New()

	vpnAddr := net.JoinHostPort("127.0.0.1", vpnListenPort)
	vpnUDPAddr, err := net.ResolveUDPAddr("udp", vpnAddr)
	if err != nil {
		logger.Error(fmt.Errorf("error resolving vpn udp addr: %v", err), nil)
		os.Exit(1)
	}

	serverAddr := net.JoinHostPort(serverIP, serverPort)
	logger.Info(fmt.Sprintf("Proxy server address: %s", serverAddr), nil)

	switch protocol {
	case "tcp":
		t := tcpTransport{
			serverAddr:    serverAddr,
			vpnListenPort: vpnListenPort,
			logger:        logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}
	case "utls":
		t := utlsTransport{
			serverAddr: serverAddr,
			vpnUDPAddr: vpnUDPAddr,
			logger:     logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}
	default:
		t := wsTransport{
			serverAddr: serverAddr,
			vpnUDPAddr: vpnUDPAddr,
			logger:     logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}
	}
}
