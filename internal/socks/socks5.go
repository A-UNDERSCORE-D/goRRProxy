package socks

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"

	"githib.com/A-UNDERSCORE-D/goRRProxy/pkg/sockUtils"
)

const (
	socksVersion = 0x05

	authNone           = 0x00 // We're only gonna accept this, but knowing what some of the rest are is nice
	authGSSAPI         = 0x01
	authUserPasswd     = 0x02
	authNoneAcceptable = 0xFF

	cmdConnect = 0x01

	addrV4     = 0x01
	addrDomain = 0x03
	addrV6     = 0x04
)

const (
	repOk              = 0x0
	repGenFail         = 0x1
	repCmdNotSupported = 0x07
)

func writeWithVersion(conn net.Conn, data []byte) error {
	toSend := append([]byte{socksVersion}, data...)
	_, err := conn.Write(toSend)
	return err
}

func readIPv4(conn net.Conn) (string, error) {
	data, err := sockUtils.ReadNBytes(conn, 4)
	if err != nil {
		return "", err
	}

	return net.IP(data).String(), nil
}

func readIPV6(conn net.Conn) (string, error) {
	data, err := sockUtils.ReadNBytes(conn, 16)
	if err != nil {
		return "", err
	}
	return net.IP(data).String(), nil
}

func readFQDN(conn net.Conn) (string, error) {
	fqdnLen, err := sockUtils.ReadOneByte(conn)
	if err != nil {
		return "", err
	}
	data, err := sockUtils.ReadNBytes(conn, int(fqdnLen))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func NegotiateSocksToConnect(conn net.Conn) (string, string, error) {
	fmt.Println("[socks] starting request")
	ver, err := sockUtils.ReadOneByte(conn)
	if err != nil {
		return "", "", err
	}
	if ver != socksVersion {
		return "", "", errors.New("invalid socks version")
	}
	fmt.Println("[socks] getting methods")
	methCount, err := sockUtils.ReadOneByte(conn)
	if err != nil {
		return "", "", err
	}
	fmt.Println("[socks] requested method num: ", int(methCount))
	methods, err := sockUtils.ReadNBytes(conn, int(methCount))
	fmt.Println("[socks] requested methods: ", methods)
	if !bytes.Contains(methods, []byte{authNone}) {
		_ = writeWithVersion(conn, []byte{authNoneAcceptable})
		_ = conn.Close()
		return "", "", errors.New("no acceptable authentication methods supplied")
	}

	writeWithVersion(conn, []byte{authNone})

	if v, err := sockUtils.ReadOneByte(conn); err != nil {
		return "", "", err
	} else if v != socksVersion {
		return "", "", errors.New("invalid socks version")
	}

	if cmd, err := sockUtils.ReadOneByte(conn); err != nil {
		return "", "", err
	} else if cmd != cmdConnect {
		_ = writeWithVersion(conn, []byte{repCmdNotSupported})
	}
	_, _ = sockUtils.ReadOneByte(conn) // eat the reserved byte
	var targetAddr string
	var addrErr error
	if addrType, err := sockUtils.ReadOneByte(conn); err != nil {
		return "", "", err
	} else {
		switch addrType {
		case addrV4:
			targetAddr, addrErr = readIPv4(conn)
		case addrV6:
			targetAddr, addrErr = readIPV6(conn)
		case addrDomain:
			targetAddr, addrErr = readFQDN(conn)
		default:
			return "", "", errors.New("invalid address type")
		}
	}

	if addrErr != nil {
		return "", "", addrErr
	}

	portBytes, err := sockUtils.ReadNBytes(conn, 2)
	if err != nil {
		return "", "", err
	}
	port := binary.BigEndian.Uint16(portBytes)

	return targetAddr, strconv.Itoa(int(port)), nil
}

var generalErrorBytes = []byte{
	0x01,                   // Error code
	0x00,                   // Reserved
	0x01,                   // IPv4 address type (dummy)
	0x00, 0x00, 0x00, 0x00, // IPv4 length stream of nulls
	0x00, 0x00,             // two bytes for the port, also dummy
}

func WriteGeneralError(conn net.Conn) error {
	return writeWithVersion(conn, generalErrorBytes)
}

func WriteConnectSuccess(conn net.Conn, bindAddr string, bindPort int) error {
	buf := new(bytes.Buffer)
	buf.WriteByte(repOk)
	buf.WriteByte(0x00) // reserved byte
	ip := net.ParseIP(bindAddr)

	switch {
	case ip == nil: // Its a host, our first byte is the length
		buf.WriteByte(addrDomain)
		l := byte(len(bindAddr))
		buf.WriteByte(l)
		buf.WriteString(bindAddr)
	case ip.To4() == nil: // To4 returns nil when the IP is not a valid v4 IP, and since we checked above, it must be a v6
		buf.WriteByte(addrV6)
		buf.Write(ip.To16())
	default: // Must be a v4 if we get here
		buf.WriteByte(addrV4)
		buf.Write(ip.To4())
	}
	_ = binary.Write(buf, binary.LittleEndian, uint16(bindPort)) // write the port in network octet order per RFC
	return writeWithVersion(conn, buf.Bytes())
}
