package connection

import (
	"math"
	"net"
	"net/netip"
	"time"

	"github.com/julioguillermo/jg_sender/config"
)

type Scanner struct {
	conf     *config.Config
	Running  bool
	Progress func(float64)
	Found    func()
}

func NewScanner(conf *config.Config) *Scanner {
	return &Scanner{
		conf:     conf,
		Running:  false,
		Progress: nil,
		Found:    nil,
	}
}

func (p *Scanner) Stop() {
	p.Running = false
}

func (p *Scanner) ScannAll(subnets []*netip.Prefix) {
	p.Running = true
	maxAddr := 0.0
	addrPro := 0.0
	for _, sn := range subnets {
		maxAddr += math.Pow(2, float64(32-sn.Bits()))
	}

	addrPro = 0
	ctl := make(chan bool, p.conf.Connections())
	end := make(chan bool)
	for _, sn := range subnets {
		if p.Running {
			addr := sn.Masked().Addr()
			for p.Running && sn.Contains(addr) {
				ctl <- true
				a := netip.AddrFrom4(addr.As4())
				if addrPro < maxAddr-1 {
					go p.scanAddr(&a, ctl, nil)
				} else {
					go p.scanAddr(&a, ctl, end)
				}
				addr = addr.Next()
				addrPro++
				if p.Progress != nil {
					p.Progress(addrPro / maxAddr)
				}
			}
		}
	}

	<-end
	close(ctl)
	close(end)
	p.Running = false
	if p.Progress != nil {
		p.Progress(-1)
	}
}

func (p *Scanner) scanAddr(a *netip.Addr, c chan bool, e chan bool) {
	addrPort := netip.AddrPortFrom(*a, uint16(config.Port))
	ctl, uuid, name, dtype := p.scanAddrPort(&addrPort)

	if ctl {
		SetDevice(uuid, &Device{
			ID:   uuid,
			Addr: a,
			Name: name,
			OS:   dtype,
		})
		if p.Found != nil {
			p.Found()
		}
	}

	if e != nil {
		e <- true
	}
	<-c
}

func (p *Scanner) scanAddrPort(addr *netip.AddrPort) (bool, string, string, string) {
	conn, err := net.DialTimeout("tcp", addr.String(), time.Duration(p.conf.Timeout())*time.Millisecond)
	if err == nil {
		defer conn.Close()
		_, err = conn.Write([]byte{NAME})
		if err != nil {
			return false, "", "", ""
		}

		bint := make([]byte, 8)
		_, err = conn.Read(bint)
		if err != nil {
			return false, "", "", ""
		}
		uuid_buf := make([]byte, BytesToInt(bint))
		_, e := conn.Read(uuid_buf)
		if e != nil {
			return false, "", "", ""
		}
		uuid := string(uuid_buf)

		_, err = conn.Read(bint)
		if err != nil {
			return false, "", "", ""
		}
		name_buf := make([]byte, BytesToInt(bint))
		_, e = conn.Read(name_buf)
		if e != nil {
			return false, "", "", ""
		}
		name := string(name_buf)

		_, err = conn.Read(bint)
		if err != nil {
			return false, "", "", ""
		}
		os_buf := make([]byte, BytesToInt(bint))
		_, e = conn.Read(os_buf)
		if e != nil {
			return false, "", "", ""
		}
		os := string(os_buf)

		return true, uuid, name, os
	}
	return false, "", "", ""
}
