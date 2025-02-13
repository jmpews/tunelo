package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"tunelo"
	"tunelo/pkg/logger/plain"
)

func main() {
	var serverIP string
	var serverPort string
	var vpnListenPort string
	var vpnEndpointPort string
	var protocol string
	var serverDomain string

	flag.StringVar(&serverIP, "server_ip", "127.0.0.1", "Remote proxy-server IP address.")
	flag.StringVar(&serverPort, "server_port", "11821", "Remote proxy-server port number.")
	flag.StringVar(&vpnEndpointPort, "vpn_endpoint_port", "11820", "VPN endpoint port number.")
	flag.StringVar(&vpnListenPort, "vpn_listen_port", "0", "VPN listen port number.")
	flag.StringVar(&protocol, "p", "ws", "Tunnel transport protocol. Options: ws, utls, and tcp.")
	flag.StringVar(&serverDomain, "server_domain", "", "Server domain.")
	flag.Parse()

	logger := plain.New()

	vpnEndpointUDPConn, err := tunelo.CreateUDPListenConn(vpnEndpointPort)
	if err != nil {
		logger.Error(fmt.Errorf("error creating vpn endpoint udp listen conn: %v", err), nil)
		os.Exit(1)
	}
	defer vpnEndpointUDPConn.Close()
	logger.Info(fmt.Sprintf("VPN endpoint udp listen conn: %s", vpnEndpointUDPConn.LocalAddr().String()), nil)

	var vpnListenUDPConn *net.UDPConn = nil
	if vpnListenPort != "0" {
		vpnListenUDPConn, err := tunelo.CreateUDPListenConn(vpnListenPort)
		if err != nil {
			logger.Error(fmt.Errorf("error creating vpn listen udp conn: %v", err), nil)
			os.Exit(1)
		}
		defer vpnListenUDPConn.Close()
		logger.Info(fmt.Sprintf("VPN listen udp conn: %s", vpnListenUDPConn.LocalAddr().String()), nil)
	}

	serverAddr := net.JoinHostPort(serverIP, serverPort)

	switch protocol {
	case "tcp":
		t := tcpTransport{
			serverAddr:         serverAddr,
			vpnListenUDPConn:   vpnListenUDPConn,
			vpnEndpointUDPConn: vpnEndpointUDPConn,
			logger:             logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}
	case "utls":
		t := utlsTransport{
			serverDomain:       serverDomain,
			serverAddr:         serverAddr,
			vpnEndpointUDPConn: vpnEndpointUDPConn,
			vpnListenUDPConn:   vpnListenUDPConn,
			logger:             logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}

	default:
		t := wsTransport{
			serverAddr:    serverAddr,
			vpnConn:       vpnListenUDPConn,
			clientUDPConn: vpnEndpointUDPConn,
			logger:        logger,
		}
		err := t.run()
		if err != nil {
			logger.Error(err, nil)
			os.Exit(1)
		}
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	<-ctx.Done()

	fmt.Println("\n[-] shutdown signal received")
	os.Exit(0)
}
