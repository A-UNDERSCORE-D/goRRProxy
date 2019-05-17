package sockUtils

import (
	"errors"
	"net"
)

var ErrNotEnoughRead = errors.New("did not read enough data from socket to satisfy request")

func ReadNBytes(conn net.Conn, numRead int) ([]byte, error) {
	out := make([]byte, numRead)
	if i, err := conn.Read(out); err != nil {
		return nil, err
	} else if i != numRead {
		return out, ErrNotEnoughRead
	}
	return out, nil
}

func ReadOneByte(conn net.Conn) (byte, error) {
	data, err := ReadNBytes(conn, 1)
	if err != nil {
		return 0x0, err
	}
	return data[0], nil
}
