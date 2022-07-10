package connection

import (
	"net"
	"net/netip"

	"github.com/julioguillermo/jg_sender/config"
)

func SendMSG(conf *config.Config, addr *netip.Addr, msg string, onError func(error)) {
	if onError == nil {
		onError = func(error) {}
	}

	addrPort := netip.AddrPortFrom(*addr, uint16(config.Port))

	conn, err := net.Dial("tcp", addrPort.String())
	if err != nil {
		onError(err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte{MSG})
	if err != nil {
		onError(err)
		return
	}

	_, err = conn.Write(IntToBytes(uint64(len(conf.Name()))))
	if err != nil {
		onError(err)
		return
	}

	_, err = conn.Write(IntToBytes(uint64(len(msg))))
	if err != nil {
		onError(err)
		return
	}

	_, err = conn.Write([]byte(conf.Name()))
	if err != nil {
		onError(err)
		return
	}

	_, err = conn.Write([]byte(msg))
	if err != nil {
		onError(err)
	}
}
