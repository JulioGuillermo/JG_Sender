package connection

import (
	"fmt"
	"net"

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
	}
}
