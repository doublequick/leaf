package network

import (
	"net"
)

type PlayerConn interface {
	ReadMsg() ([]byte, error)
	WriteMsg(args ...[]byte) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
}
