package connection

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/julioguillermo/jg_sender/config"
	"github.com/julioguillermo/jg_sender/config/storage"
)

var Serv *Server

type Server struct {
	conf   *config.Config
	Serv   net.Listener
	OnFile func(addr, user, files string, onCancel func()) (func(uint64, uint64, uint64, uint64), func(error))
	OnMSG  func(addr, name, msg string)
}

func InitServer(conf *config.Config) *Server {
	server, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprint(config.Port))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	Serv = &Server{
		conf:   conf,
		Serv:   server,
		OnFile: nil,
		OnMSG:  nil,
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
		name := p.conf.Name()
		os := p.conf.OS()
		connection.Write(IntToBytes(uint64(len(name))))
		connection.Write(IntToBytes(uint64(len(os))))
		connection.Write([]byte(name))
		connection.Write([]byte(os))
	case MSG:
		p.GetMSG(connection)
	case RESOURCES:
		p.GetResources(connection)
	case EXPLORE:
		p.explore(connection)
	case GET:
		p.send(connection)
	}
}

func (p *Server) GetMSG(connection net.Conn) {
	bulen := make([]byte, 8)
	bmlen := make([]byte, 8)
	_, e := connection.Read(bulen)
	if e != nil {
		return
	}
	_, e = connection.Read(bmlen)
	if e != nil {
		return
	}

	user_len := BytesToInt(bulen)
	msg_len := BytesToInt(bmlen)

	buser := make([]byte, user_len)
	bmsg := make([]byte, msg_len)

	_, e = connection.Read(buser)
	if e != nil {
		return
	}
	_, e = connection.Read(bmsg)
	if e != nil {
		return
	}

	if p.OnMSG != nil {
		p.OnMSG(connection.RemoteAddr().String(), string(buser), string(bmsg))
	}
}

func FileExist(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

/**
Recive:
-Ctl byte: RESOURCES
-len(UserName)
-UserName
-BufSize
-TSize
-len(files)
for f -> files:
	-len(fileName)
	-fileName
for f -> files:
	-size(f)
	for pro < size(f):
		-[256]bytes(f)
*/
func (p *Server) GetResources(connection net.Conn) {
	var err error
	bint := make([]byte, 8)
	// User
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	userName := make([]byte, BytesToInt(bint))
	_, err = connection.Read(userName)
	if err != nil {
		return
	}
	// BufSize, total size and number of files
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	BufSize := BytesToInt(bint)
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	t_size := BytesToInt(bint)
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	n_files := BytesToInt(bint)

	// files names
	files := make([]string, n_files)
	name := ""
	var fExt string
	var fPath string
	var fName string
	var fileName []byte
	var num uint64
	for i := range files {
		_, err = connection.Read(bint)
		if err != nil {
			return
		}
		fileName = make([]byte, BytesToInt(bint))
		_, err = connection.Read(fileName)
		if err != nil {
			return
		}
		fName = string(fileName)
		fExt = path.Ext(fName)
		fName = fName[:len(fName)-len(fExt)]
		fPath = fName + fExt
		num = 0
		for FileExist(path.Join(p.conf.Inbox(), fPath)) {
			num++
			fPath = fmt.Sprintf("%s_(%d)%s", fName, num, fExt)
		}
		files[i] = fPath
		name += fPath + "\n"
	}

	var (
		onProgress func(uint64, uint64, uint64, uint64)
		onError    func(error)
	)
	if p.OnFile != nil {
		onProgress, onError = p.OnFile(connection.RemoteAddr().String(), string(userName), name, func() {
			connection.Close()
		})
	} else {
		onProgress = func(uint64, uint64, uint64, uint64) {}
		onError = func(error) {}
	}
	// Recive files
	var fp string
	var dir string
	var file *os.File
	var pro uint64
	var size uint64
	var t int
	buf := make([]byte, BufSize)
	t_pro := uint64(0)
	for i, f := range files {
		fp = path.Join(p.conf.C_InboxDir, f)
		dir = path.Dir(fp)
		if dir != "." {
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				onError(err)
				return
			}
		}
		file, err = os.Create(fp)
		if err != nil {
			onError(err)
			return
		}
		pro = 0
		_, err = connection.Read(bint)
		if err != nil {
			os.Remove(fp)
			file.Close()
			onError(err)
			return
		}
		size = BytesToInt(bint)
		for pro < size {
			if size-pro < BufSize {
				t, err = connection.Read(buf[:size-pro])
			} else {
				t, err = connection.Read(buf)
			}
			if err != nil {
				os.Remove(fp)
				file.Close()
				onError(err)
				return
			}
			_, err = file.Write(buf[:t])
			if err != nil {
				os.Remove(fp)
				file.Close()
				onError(err)
				return
			}
			pro += uint64(t)
			t_pro += uint64(t)
			onProgress(uint64(i), uint64(len(files)), t_pro, t_size)
		}
		file.Close()
	}
	onProgress(uint64(len(files)), uint64(len(files)), t_size, t_size)
}

/*
Recive:
	CTL
	path_size
	path
Send:
	if Error:
		ERROR_CTL and error
	else:
		EXPLORE_CTL
		Number of elements
		for elements:
			is dir
			path len
			path
*/
func (p *Server) explore(connection net.Conn) {
	var err error
	// CTL
	ctl := make([]byte, 8)
	_, err = connection.Read(ctl)
	if err != nil {
		return
	}
	if !CheckCTL(ctl) {
		return
	}
	// Get path len and path
	p_size := make([]byte, 8)
	_, err = connection.Read(p_size)
	if err != nil {
		return
	}
	bpath := make([]byte, BytesToInt(p_size))
	_, err = connection.Read(bpath)
	if err != nil {
		return
	}

	// Read path
	elements, err := storage.Explore(string(bpath))
	// IF err: send error
	if err != nil {
		e := err.Error()
		_, err = connection.Write([]byte{ERROR})
		if err != nil {
			return
		}
		_, err = connection.Write(IntToBytes(uint64(len(e))))
		if err != nil {
			return
		}
		connection.Write([]byte(e))
		return
	}

	// Not error: send ctl -> EXPLORE
	_, err = connection.Write([]byte{EXPLORE})
	if err != nil {
		return
	}
	// Send number of elements
	_, err = connection.Write(IntToBytes(uint64(len(elements))))
	if err != nil {
		return
	}

	// For each element
	for _, e := range elements {
		// Send if is dir or file
		if e.IsDir {
			_, err = connection.Write([]byte{DIR})
			if err != nil {
				return
			}
		} else {
			_, err = connection.Write([]byte{FILE})
			if err != nil {
				return
			}
		}
		// Send path len and path
		_, err = connection.Write(IntToBytes(uint64(len(e.Path))))
		if err != nil {
			return
		}
		_, err = connection.Write([]byte(e.Path))
		if err != nil {
			return
		}
	}
}

/*
Recive:
	CTL
	num of path
	for each path:
		path size
		path
	send all path width sender
*/
func (p *Server) send(connection net.Conn) {
	var err error
	// CTL
	ctl := make([]byte, len(CTL))
	_, err = connection.Read(ctl)
	if err != nil {
		return
	}
	if !CheckCTL(ctl) {
		return
	}

	bint := make([]byte, 8)
	// Get num of path
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	// Get paths
	paths := make([]string, BytesToInt(bint))
	var bufpath []byte
	for i := range paths {
		// Get path len and path
		_, err = connection.Read(bint)
		if err != nil {
			return
		}
		bufpath = make([]byte, BytesToInt(bint))
		_, err = connection.Read(bufpath)
		if err != nil {
			return
		}
		paths[i] = string(bufpath)
	}

	// send
	sender := NewSender()
	sender.connection = connection
	sender.sendResources(p.conf, paths, nil, nil)
}
