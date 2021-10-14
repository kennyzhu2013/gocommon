// +build linux

/*
@Time : 2019/12/10 17:28
@Author : kenny zhu
@File : re-use port for linux
@Software: GoLand
@Others:
*/
package reuse

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// See net.RawConn.Control
func Control(network, address string, c syscall.RawConn) (err error) {
	_ = c.Control(func(fd uintptr) {

		if err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			return 
		}
		if err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			return
		}
	})
	return nil
}

var (
	listenConfig = net.ListenConfig{
		Control: Control,
	}
)

// Listen tcp with SO_REUSEPORT and SO_REUSEADDR option set.
func Listen(network, address string) (net.Listener, error) {
	return listenConfig.Listen(context.Background(), network, address)
}

// Listen udp with SO_REUSEPORT and SO_REUSEADDR option set.
func ListenPacket(network, address string) (net.PacketConn, error) {
	return listenConfig.ListenPacket(context.Background(), network, address)
}

// Dial dials the given network and address with SO_REUSEPORT and SO_REUSEADDR option set.
func Dial(network, laddr, raddr string) (net.Conn, error) {
	nla, err := ResolveAddr(network, laddr)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{
		Control:   Control,
		LocalAddr: nla,
	}

	return d.Dial(network, raddr)
}

func ResolveAddr(network, address string) (net.Addr, error) {
	switch network {
	case "ip", "ip4", "ip6":
		return net.ResolveIPAddr(network, address)
	case "tcp", "tcp4", "tcp6":
		return net.ResolveTCPAddr(network, address)
	case "udp", "udp4", "udp6":
		return net.ResolveUDPAddr(network, address)
	case "unix", "unixgram", "unixpacket":
		return net.ResolveUnixAddr(network, address)
	default:
		return nil, net.UnknownNetworkError(network)
	}
}
