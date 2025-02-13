package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"tunelo"

	"tunelo/pkg/logger"
)

type tcpTransport struct {
	serverAddr         string
	vpnEndpointUDPConn *net.UDPConn
	vpnListenUDPConn   *net.UDPConn
	logger             logger.Logger
}

func (t *tcpTransport) run() error {
	host, port, err := net.SplitHostPort(t.serverAddr)
	tcpConn, err := tunelo.ConnectTCPConn(host, port)
	if err != nil {
		return fmt.Errorf("error dialling tcp server: %v", err)
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
			fmt.Printf("error closing tcp conn: %v", err)
		}
		fmt.Printf("tcp connection %s closed\n", tcpConn.RemoteAddr().String())
	}(tcpConn)
	t.logger.Info(fmt.Sprintf("Server TCP conn: %s", tcpConn.RemoteAddr().String()), nil)

	t.logger.Info("tcp connected. Tunneling...", nil)

	if t.vpnListenUDPConn == nil {
		buf := make([]byte, 1024)
		n, vpnListenAddr, err := t.vpnEndpointUDPConn.ReadFromUDP(buf)
		if err != nil {
			return fmt.Errorf("error reading from vpn endpoint udp conn: %v", err)
		}

		_, err = tcpConn.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("error writing to tcp conn: %v", err)
		}

		// create vpn listen udp conn
		t.vpnListenUDPConn, err = tunelo.ConnectUDPConn(vpnListenAddr.IP.String(), fmt.Sprintf("%d", vpnListenAddr.Port))
		if err != nil {
			return fmt.Errorf("error creating vpn listen udp conn: %v", err)
		}
		defer func(c *net.UDPConn) {
			err := c.Close()
			if err != nil {
				fmt.Printf("error closing vpn listen udp conn: %v", err)
			}
			fmt.Printf("vpn listen udp conn %s closed\n", t.vpnListenUDPConn.LocalAddr().String())
		}(t.vpnListenUDPConn)
		t.logger.Info(fmt.Sprintf("VPN listen UDP conn: %s", t.vpnListenUDPConn.LocalAddr().String()), nil)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		_, err := io.Copy(tcpConn, t.vpnEndpointUDPConn)
		if err != nil {
			t.logger.Error(fmt.Errorf("error copying from vpn endpoint udp conn to tcp conn: %v", err), nil)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(t.vpnListenUDPConn, tcpConn)
		if err != nil {
			t.logger.Error(fmt.Errorf("error copying from tcp conn to vpn endpoint udp conn: %v", err), nil)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(tcpConn, t.vpnListenUDPConn)
		if err != nil {
			t.logger.Error(fmt.Errorf("error copying from vpn listen udp conn to tcp conn: %v", err), nil)
		}
	}()

	wg.Wait()

	return nil
}
