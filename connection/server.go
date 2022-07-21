package connection

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/julioguillermo/jg_sender/config"
)

var Serv *Server

type Server struct {
	conf          *config.Config
	Serv          net.Listener
	UpdateHistory func(UserID string)
	Notify        func(UserID, title, txt string)
}

func InitServer(conf *config.Config) *Server {
	server, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprint(config.Port))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	Serv = &Server{
		conf: conf,
		Serv: server,
	}
	go Serv.ProcessServer()
	return Serv
}

func (p *Server) ProcessServer() {
	for {
		connection, err := p.Serv.Accept()
		if err == nil {
			go p.ProcessClient(connection)
		}
	}
}

func (p *Server) ProcessClient(connection net.Conn) {
	defer connection.Close()
	ctl := make([]byte, 1)
	_, e := connection.Read(ctl)
	if e != nil || len(ctl) == 0 {
		return
	}
	switch ctl[0] {
	case NAME:
		uuid := p.conf.UUID
		name := p.conf.Name()
		os := p.conf.OS()
		connection.Write(IntToBytes(uint64(len(uuid))))
		connection.Write([]byte(uuid))
		connection.Write(IntToBytes(uint64(len(name))))
		connection.Write([]byte(name))
		connection.Write(IntToBytes(uint64(len(os))))
		connection.Write([]byte(os))
	case MSG:
		p.GetMSG(connection)
	case RESOURCES:
		p.GetResources(connection)
	case CONT_TRANS:
		p.ContinueSendingTrans(connection)
	case USER_VIEW:
		p.UserView(connection)
	}
}

func (p *Server) ContinueTrans(userID string, trans *Transfer) {
	trans.Error = nil
	trans.File.Canceled = false
	p.UpdateHistory(userID)
	if trans.In {
		p.ContinueRecivingTrans(userID, trans)
	} else {
		p.SendTrans(userID, trans)
	}
}

func (p *Server) UserView(connection net.Conn) {
	bint := make([]byte, 8)

	// User ID
	_, e := connection.Read(bint)
	if e != nil {
		return
	}
	buserID := make([]byte, BytesToInt(bint))
	_, e = connection.Read(buserID)
	if e != nil {
		return
	}
	UserID := string(buserID)

	UserView(UserID)
	if p.UpdateHistory != nil {
		p.UpdateHistory(UserID)
	}
}

func (p *Server) SendUserView(userID string) {
	dev := GetDevice(userID)
	if dev == nil {
		return
	}

	addrPort := netip.AddrPortFrom(*dev.Addr, uint16(config.Port))
	connection, e := net.Dial("tcp", addrPort.String())
	if e != nil {
		return
	}

	_, e = connection.Write([]byte{USER_VIEW})
	if e != nil {
		return
	}

	// User ID
	_, e = connection.Write(IntToBytes(uint64(len(userID))))
	if e != nil {
		return
	}
	_, e = connection.Write([]byte(userID))
	if e != nil {
		return
	}
}
