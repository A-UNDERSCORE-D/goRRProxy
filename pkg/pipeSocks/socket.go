package pipeSocks

// idea adapted from https://gist.github.com/jbardin/821d08cb64c01c84b81a

import (
	"fmt"
	"io"
	"net"
)

// ProxySockets creates a two way proxy between two sockets
func ProxySockets(sock1, sock2 net.Conn) {
	sock1Closed, sock2Closed := make(chan struct{}), make(chan struct{})
	go proxyOneWay(sock1, sock2, sock1Closed)
	go proxyOneWay(sock2, sock1, sock2Closed)

	var waitForClose chan struct{}
	select {
	case <-sock1Closed:
		waitForClose = sock2Closed
		if err := sock2.Close(); err != nil {
			fmt.Printf("error closing sock2: %s\n", err)
		}
	case <-sock2Closed:
		waitForClose = sock1Closed
		if err := sock1.Close(); err != nil {
			fmt.Printf("error closing sock1: %s\n", err)
		}
	}
	<-waitForClose
}

func proxyOneWay(src, dst net.Conn, srcClosed chan struct{}) {
	if _, err := io.Copy(src, dst); err != nil {
		fmt.Printf("error from copy: %s\n", err)
	}
	if err := src.Close(); err != nil {
		fmt.Printf("error closing source socket: %s\n", err)
	}
	srcClosed <- struct{}{}
}
