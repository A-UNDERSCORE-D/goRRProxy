package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
)

func ReadRequest(sock net.Conn) ([]byte, error) {
	dataSizeBytes := make([]byte, 2)
	i, err := sock.Read(dataSizeBytes)
	if err != nil {
		return nil, err
	}
	if i != 2 {
		return nil, errors.New("did not read enough from socket")
	}

	dataSize := binary.BigEndian.Uint16(dataSizeBytes)
	out := make([]byte, dataSize)
	i, err = sock.Read(out)
	if err != nil {
		return nil, err
	}
	if i != int(dataSize) {
		return nil, fmt.Errorf("did not read enough from socket: got %d, wanted %d. Leftover: %#v", i, dataSize, out[:i])
	}

	return out, nil
}

func WriteRequest(data []byte, sock net.Conn) error {
	if len(data) > math.MaxUint16 {
		return errors.New("payload too large")
	}
	toWrite := bytes.Buffer{}
	if err := binary.Write(&toWrite, binary.BigEndian, uint16(len(data))); err != nil {
		return err
	}

	toWrite.Write(data)
	l := toWrite.Len()
	i, err := sock.Write(toWrite.Bytes())
	if err != nil {
		return err
	}
	if i != l {
		return errors.New("did not write all data")
	}
	return nil
}

func WriteJson(data interface{}, sock net.Conn) error {
	toSend, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return WriteRequest(toSend, sock)
}

func ReadJson(data interface{}, sock net.Conn) error {
	read, err := ReadRequest(sock)
	if err != nil {
		return err
	}
	return json.Unmarshal(read, data)
}
