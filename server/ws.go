package main

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"nhooyr.io/websocket"

	"tunelo/pkg/logger"
)

type wsTransport struct {
	serverAddr string
	vpnUDPAddr *net.UDPAddr
	logger     logger.Logger
}

func (t *wsTransport) run() error {
	http.HandleFunc("/ws", t.handler)

	t.logger.Info(fmt.Sprintf("WebSocket server listening on %s", t.serverAddr), nil)

	err := http.ListenAndServe(t.serverAddr, nil)
	if err != nil {
		return fmt.Errorf("error listening: %v", err)
	}

	return nil
}

func (t *wsTransport) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		t.logger.Error(fmt.Errorf("error accepting ws conn: %v", err), nil)
		return
	}
	defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
		err := conn.Close(code, reason)
		if err != nil {
			t.logger.Error(fmt.Errorf("error closing conn: %v", err), nil)
		}
	}(conn, websocket.StatusNormalClosure, "")

	wsNetConn := websocket.NetConn(r.Context(), conn, websocket.MessageBinary)

	t.logger.Info("ws connection accepted. Tunneling...", nil)

	localUDPAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
	if err != nil {
		t.logger.Error(fmt.Errorf("error resolving local udp addr: %v", err), nil)
		return
	}

	vpnConn, err := net.DialUDP("udp", localUDPAddr, t.vpnUDPAddr)
	if err != nil {
		t.logger.Error(fmt.Errorf("error dialling vpn: %v", err), nil)
		return
	}
	defer func(c *net.UDPConn) {
		err := c.Close()
		if err != nil {
			t.logger.Error(fmt.Errorf("error closing udp conn: %v", err), nil)
		}
	}(vpnConn)
	t.logger.Info("udp connection established. Tunneling...", nil)

	go t.handle(wsNetConn, vpnConn)
}

func (t *wsTransport) handle(conn net.Conn, vpnConn *net.UDPConn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			t.logger.Error(fmt.Errorf("error closing ws conn: %v", err), nil)
		}
	}(conn)

	go func() {
		_, err := io.Copy(vpnConn, conn)
		if err != nil {
			t.logger.Error(fmt.Errorf("error copying from ws conn to vpn: %v", err), nil)
		}
	}()

	_, err := io.Copy(conn, vpnConn)
	if err != nil {
		t.logger.Error(fmt.Errorf("error copying from vpn to ws conn: %v", err), nil)
	}
}
