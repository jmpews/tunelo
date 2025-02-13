package tunelo

import (
	"net"
)

func CreateUDPListenConn(port string) (*net.UDPConn, error) {
	addr := net.JoinHostPort("127.0.0.1", port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	listenUDPConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return listenUDPConn, nil
}

func ConnectUDPConn(host, port string) (*net.UDPConn, error) {
	addr := net.JoinHostPort(host, port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	localUDPAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}

	udpConn, err := net.DialUDP("udp", localUDPAddr, udpAddr)
	if err != nil {
		return nil, err
	}

	return udpConn, nil
}

func ConnectTCPConn(host, port string) (*net.TCPConn, error) {
	addr := net.JoinHostPort(host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}

func CreateTCPListenConn(port string) (*net.TCPListener, error) {
	addr := net.JoinHostPort("0.0.0.0", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	listenTCPConn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	return listenTCPConn, nil
}
