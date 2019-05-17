package client

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/protocol"
	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/request"
	"githib.com/A-UNDERSCORE-D/goRRProxy/pkg/pipeSocks"
)

func ListenForConnections(bindAddr, bindPort string) error {
	l, err := net.Listen("tcp", net.JoinHostPort(bindAddr, bindPort))
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("got an error from the listener: ", err)
		}

		go handleConn(conn)

	}
}

func handleConn(conn net.Conn) {
	var req request.ProxyRequest
	fmt.Println("parsing json payload")
	if err := protocol.ReadJson(&req, conn); err != nil {
		fmt.Println("error while parsing JSON payload: ", err)
		return
	}
	fmt.Println(req)
	// TODO: Validate API keys here
	fmt.Println("dialing target: ", req.Target, req.TargetPort)
	sock, err := net.Dial("tcp", net.JoinHostPort(req.Target, req.TargetPort))
	if err != nil {
		_ = protocol.WriteJson(request.ProxyResponse{Ok: false, Err: err.Error()}, conn)
		return
	}
	fmt.Println("dial complete")
	host, portStr, _ := net.SplitHostPort(sock.LocalAddr().String())
	portNum, _ := strconv.Atoi(portStr) // we can be sure this is valid as it came from a live socket
	fmt.Println("writing json payload")
	_ = protocol.WriteJson(request.ProxyResponse{Ok:true, BindAddr:host, BindPort:portNum}, conn)
	fmt.Println("sleeping")
	time.Sleep(time.Second)
	fmt.Println("proxying sockets")
	pipeSocks.ProxySockets(conn, sock)
}
