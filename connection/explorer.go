package connection

import (
	"errors"
	"net"
	"net/netip"

	"github.com/julioguillermo/jg_sender/config"
)

type Explorer struct {
	connection net.Conn
}

func NewExplorer() *Explorer {
	exp := &Explorer{}
	return exp
}

func (p *Explorer) Explore(addr *netip.Addr, pth string, onError func(error), onPath func(bool, string, uint64, uint64)) {
	if onError == nil {
		onError = func(error) {}
	}
	if onPath == nil {
		onPath = func(bool, string, uint64, uint64) {}
	}

	var err error
	// Open connection
	addrPort := netip.AddrPortFrom(*addr, uint16(config.Port))
	p.connection, err = net.Dial("tcp", addrPort.String())
	if err != nil {
		onError(err)
		return
	}

	// Notify that we want to explore
	_, err = p.connection.Write([]byte{EXPLORE})
	if err != nil {
		onError(err)
		return
	}

	// Send ctl
	_, err = p.connection.Write(CTL)
	if err != nil {
		onError(err)
		return
	}

	// Send path len and path
	_, err = p.connection.Write(IntToBytes(uint64(len(pth))))
	if err != nil {
		onError(err)
		return
	}
	_, err = p.connection.Write([]byte(pth))
	if err != nil {
		onError(err)
		return
	}

	bint := make([]byte, 8)
	// Check for errors
	ctl := make([]byte, 1)
	_, err = p.connection.Read(ctl)
	if err != nil {
		onError(err)
		return
	}
	if ctl[0] == ERROR {
		// Err size
		_, err = p.connection.Read(bint)
		if err != nil {
			onError(err)
			return
		}
		errmsg := make([]byte, BytesToInt(bint))
		_, err = p.connection.Read(errmsg)
		if err != nil {
			onError(err)
			return
		}
		onError(errors.New(string(errmsg)))
		return
	}

	// Get number of path
	_, err = p.connection.Read(bint)
	if err != nil {
		onError(err)
		return
	}
	n_path := BytesToInt(bint)

	// Get paths
	pro := uint64(0)
	var isDir bool
	var bpath []byte
	for pro < n_path {
		pro++
		// Check if path is a dir
		_, err = p.connection.Read(ctl)
		if err != nil {
			onError(err)
			return
		}
		isDir = ctl[0] == DIR
		// Path len
		_, err = p.connection.Read(bint)
		if err != nil {
			onError(err)
			return
		}
		// path
		bpath = make([]byte, BytesToInt(bint))
		_, err = p.connection.Read(bpath)
		if err != nil {
			onError(err)
			return
		}
		onPath(isDir, string(bpath), pro, n_path)
	}
}

func (p *Explorer) Download(addr *netip.Addr, elements []string, onError func(error), onStart func()) {
	if onError == nil {
		onError = func(err error) {}
	}

	var err error
	// Open connection
	addrPort := netip.AddrPortFrom(*addr, uint16(config.Port))
	p.connection, err = net.Dial("tcp", addrPort.String())
	if err != nil {
		onError(err)
		return
	}

	// Notify that we want to explore
	_, err = p.connection.Write([]byte{GET})
	if err != nil {
		onError(err)
		return
	}

	// Send ctl
	_, err = p.connection.Write(CTL)
	if err != nil {
		onError(err)
		return
	}

	// Send num of path
	_, err = p.connection.Write(IntToBytes(uint64(len(elements))))
	if err != nil {
		onError(err)
		return
	}
	// Send path len and path for each element
	for _, e := range elements {
		_, err = p.connection.Write(IntToBytes(uint64(len(e))))
		if err != nil {
			onError(err)
			return
		}
		_, err = p.connection.Write([]byte(e))
		if err != nil {
			onError(err)
			return
		}
	}

	if onStart != nil {
		onStart()
	}

	Serv.GetResources(p.connection)
}
