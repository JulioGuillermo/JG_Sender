package connection

import (
	"errors"
	"io"
	"net"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/julioguillermo/jg_sender/config"
)

type CMD struct {
	connection net.Conn

	OnError  func(error)
	OnResult func(string)
}

func NewExecutor(conn net.Conn) {
	var (
		cmd   *CMD
		onErr func(error)
		onCMD func(string)
	)
	onErr = func(err error) {
		//cmd.Close()
	}
	onCMD = func(s string) {
		defer cmd.setOnRes(onCMD)
		if len(s) > 0 && s[len(s)-1] == '\n' {
			s = s[:len(s)-1]
		}
		ca := strings.Split(s, " ")
		c := exec.Command(ca[0], ca[1:]...)
		out, err := c.StdoutPipe()
		if err != nil {
			cmd.SendCMD("ERROR: " + err.Error() + "\n")
			onErr(err)
			return
		}
		eout, err := c.StderrPipe()
		if err != nil {
			cmd.SendCMD("ERROR: " + err.Error() + "\n")
			onErr(err)
			return
		}
		in, err := c.StdinPipe()
		if err != nil {
			cmd.SendCMD("ERROR: " + err.Error() + "\n")
			onErr(err)
			return
		}

		go func() {
			buf := make([]byte, 1024)
			for {
				t, err := out.Read(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						cmd.SendCMD("ERROR (out): " + err.Error() + "\n")
						onErr(err)
					}
					return
				}
				cmd.SendCMD(string(buf[:t]))
			}
		}()
		go func() {
			buf := make([]byte, 1024)
			for {
				t, err := eout.Read(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						cmd.SendCMD("ERROR (err): " + err.Error() + "\n")
						onErr(err)
					}
					return
				}
				cmd.SendCMD(string(buf[:t]))
			}
		}()
		cmd.OnResult = func(s string) {
			_, err := in.Write([]byte(s))
			if err != nil {
				cmd.SendCMD("ERROR (in): " + err.Error() + "\n")
				onErr(err)
			}
		}

		err = c.Run()
		if err != nil {
			cmd.SendCMD("ERROR: " + err.Error() + "\n")
			onErr(err)
			return
		}
	}
	cmd = NewCMDFromConnection(conn, onErr, onCMD)
	cmd.initListener()
}

func NewCMD(addr *netip.Addr, onError func(error), onResult func(string)) *CMD {
	addrPort := netip.AddrPortFrom(*addr, uint16(config.Port))
	connection, err := net.Dial("tcp", addrPort.String())
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return nil
	}
	_, err = connection.Write([]byte{EXEC_CMD})
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return nil
	}
	_, err = connection.Write([]byte(CTL))
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return nil
	}
	cmd := NewCMDFromConnection(connection, onError, onResult)
	go cmd.initListener()
	return cmd
}

func NewCMDFromConnection(conn net.Conn, onError func(error), onResult func(string)) *CMD {
	if onError == nil {
		onError = func(error) {}
	}
	if onResult == nil {
		onResult = func(string) {}
	}
	cmd := &CMD{
		connection: conn,
		OnError:    onError,
		OnResult:   onResult,
	}
	return cmd
}

func (p *CMD) initListener() {
	defer p.Close()

	var err error
	var buf []byte
	var t int
	bint := make([]byte, 8)

	for {
		_, err = p.connection.Read(bint)
		if err != nil {
			p.OnError(err)
			return
		}
		buf = make([]byte, BytesToInt(bint))
		t, err = p.connection.Read(buf)
		if err != nil {
			p.OnError(err)
			return
		}
		go p.OnResult(string(buf[:t]))
	}
}

func (p *CMD) Close() {
	p.connection.Close()
}

func (p *CMD) SendCMD(cmd string) {
	_, err := p.connection.Write(IntToBytes(uint64(len(cmd))))
	if err != nil && p.OnError != nil {
		p.OnError(err)
	}
	_, err = p.connection.Write([]byte(cmd))
	if err != nil && p.OnError != nil {
		p.OnError(err)
	}
}

func (p *CMD) setOnRes(onRes func(string)) {
	p.OnResult = onRes
}
