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
	Found    func(a *netip.Addr, name, device string)
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
	ctl, name, dtype := p.scanAddrPort(&addrPort)

	if ctl && p.Found != nil {
		p.Found(a, name, dtype)
	}

	if e != nil {
		e <- true
	}
	<-c
}

func (p *Scanner) scanAddrPort(addr *netip.AddrPort) (bool, string, string) {
	conn, err := net.DialTimeout("tcp", addr.String(), time.Duration(p.conf.Timeout())*time.Millisecond)
	if err == nil {
		defer conn.Close()
		_, err = conn.Write([]byte{NAME})
		if err != nil {
			return false, "", ""
		}

		name_size := make([]byte, 8)
		os_size := make([]byte, 8)
		_, err = conn.Read(name_size)
		if err != nil {
			return false, "", ""
		}
		_, err = conn.Read(os_size)
		if err != nil {
			return false, "", ""
		}

		name_buf := make([]byte, BytesToInt(name_size))
		os_buf := make([]byte, BytesToInt(os_size))

		_, e := conn.Read(name_buf)
		if e != nil {
			return false, "", ""
		}
		name := string(name_buf)

		_, e = conn.Read(os_buf)
		if e != nil {
			return false, name, ""
		}
		os := string(os_buf)

		return true, name, os
	}
	return false, "", ""
}
