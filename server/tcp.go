package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"tunelo"

	"tunelo/pkg/logger"
)

type tcpTransport struct {
	serverAddr    string
	vpnListenPort string
	logger        logger.Logger
}

func (t *tcpTransport) run() error {
	_, port, _ := net.SplitHostPort(t.serverAddr)
	tcpListener, err := tunelo.CreateTCPListenConn(port)
	if err != nil {
		return fmt.Errorf("error creating tcp listener: %v", err)
	}
	defer tcpListener.Close()
	t.logger.Info(fmt.Sprintf("TCP server listening on %s", tcpListener.Addr().String()), nil)

	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			t.logger.Error(fmt.Errorf("error accepting tcp conn: %v", err), nil)
			os.Exit(1)
		}
		t.logger.Info("tcp connection accepted. Tunneling...", nil)

		vpnUDPConn, err := tunelo.ConnectUDPConn("127.0.0.1", t.vpnListenPort)
		if err != nil {
			t.logger.Error(fmt.Errorf("error dialling vpn: %v", err), nil)
			continue
		}
		defer func(c *net.UDPConn) {
			err := c.Close()
			if err != nil {
				fmt.Printf("error closing udp conn: %v", err)
			}
			fmt.Printf("udp connection %s closed\n", vpnUDPConn.LocalAddr().String())
		}(vpnUDPConn)
		t.logger.Info("udp connection established. Tunneling...", nil)

		go t.handle(tcpConn, vpnUDPConn)
	}
}

func (t *tcpTransport) handle(conn net.Conn, vpnConn *net.UDPConn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			t.logger.Error(fmt.Errorf("error closing tcp conn: %v", err), nil)
		}
	}(conn)

	go func() {
		_, err := io.Copy(vpnConn, conn)
		if err != nil {
			t.logger.Error(fmt.Errorf("error copying from tcp conn to vpn: %v", err), nil)
		}
	}()

	_, err := io.Copy(conn, vpnConn)
	if err != nil {
		t.logger.Error(fmt.Errorf("error copying from vpn to tcp conn: %v", err), nil)
	}
}
