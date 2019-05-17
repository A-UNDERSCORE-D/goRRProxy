package server

import (
	"fmt"
	"net"

	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/protocol"
	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/request"
	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/socks"
	"githib.com/A-UNDERSCORE-D/goRRProxy/pkg/pipeSocks"
)

func ListenForRequests(apiKey string, clients []string, bindAddr, bindPort string) error {
	l, err := net.Listen("tcp", net.JoinHostPort(bindAddr, bindPort))
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("errored while accepting a conn: ", err)
			continue
		}
		go handleConn(conn, apiKey, clients)
	}
}

var i = 0

func selectClient(clients []string) string {
	out := clients[i%len(clients)]
	i++
	return out
}

func handleConn(conn net.Conn, apiKey string, clients []string) {
	var err error
	fmt.Println("negotiating SOCKS with request")
	targetHost, targetPort, err := socks.NegotiateSocksToConnect(conn)
	if err != nil {
		fmt.Println("error occurred while negotiating SOCKS: ", err)
		conn.Close()
		return
	}

	var clientSock net.Conn
	tries := 0
	for {
		fmt.Println("attempting to connect to client")
		clientSock, err = net.Dial("tcp", selectClient(clients))
		if err == nil {
			break
		}
		tries++
		if tries > 10 {
			fmt.Println("client retry max reached, aborting")
			socks.WriteGeneralError(conn)
			conn.Close() // give up
			return
		}
	}
	req := request.ProxyRequest{APIKey: apiKey, Target: targetHost, TargetPort: targetPort}
	fmt.Println("writing json payload")
	if err := protocol.WriteJson(req, clientSock); err != nil {
		fmt.Println("could not send request to client: ", err)
		socks.WriteGeneralError(conn)
		conn.Close()
		clientSock.Close()
	}

	var response request.ProxyResponse
	fmt.Println("parsing json proxy response")
	if err := protocol.ReadJson(&response, clientSock); err != nil {
		fmt.Println("error while parsing response: ", err)
		socks.WriteGeneralError(conn)
		conn.Close()
		clientSock.Close()
		return
	}

	if !response.Ok {
		fmt.Println("error while attempting proxy: ", response.Err)
		socks.WriteGeneralError(conn)
		conn.Close()
		clientSock.Close()
	}
	fmt.Println("response is okay")
	fmt.Println("writing SOCKS success")
	socks.WriteConnectSuccess(conn, response.BindAddr, response.BindPort)
	fmt.Println("proxying sockets")
	pipeSocks.ProxySockets(clientSock, conn)

}
