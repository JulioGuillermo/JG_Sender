package connection

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path"
	"strings"
	"time"
)

func (p *Server) GetUser(connection net.Conn) (userID, userName, userOS, transID string, e error) {
	bint := make([]byte, 8)

	// User ID
	_, e = connection.Read(bint)
	if e != nil {
		return
	}
	buserID := make([]byte, BytesToInt(bint))
	_, e = connection.Read(buserID)
	if e != nil {
		return
	}
	userID = string(buserID)

	// User name
	_, e = connection.Read(bint)
	if e != nil {
		return
	}
	buserName := make([]byte, BytesToInt(bint))
	_, e = connection.Read(buserName)
	if e != nil {
		return
	}
	userName = string(buserName)

	// User OS
	_, e = connection.Read(bint)
	if e != nil {
		return
	}
	buserOS := make([]byte, BytesToInt(bint))
	_, e = connection.Read(buserOS)
	if e != nil {
		return
	}
	userOS = string(buserOS)

	// Update users
	addr, _ := netip.ParseAddr(strings.Split(connection.RemoteAddr().String(), ":")[0])
	SetDevice(userID, &Device{
		ID:   userID,
		Addr: &addr,
		Name: userName,
		OS:   userOS,
	})

	// Read TransID
	_, e = connection.Read(bint)
	if e != nil {
		return
	}
	bTransID := make([]byte, BytesToInt(bint))
	_, e = connection.Read(bTransID)
	if e != nil {
		return
	}
	transID = "R" + string(bTransID)
	return
}

func (p *Server) GetMSG(connection net.Conn) {
	userID, userName, _, transID, e := p.GetUser(connection)
	if e != nil {
		return
	}

	bint := make([]byte, 8)
	// MSG
	_, e = connection.Read(bint)
	if e == nil {
		bmsg := make([]byte, BytesToInt(bint))
		_, e = connection.Read(bmsg)
		if e == nil {
			SetTrans(transID, &Transfer{
				ID:       transID,
				UserID:   userID,
				DateTime: time.Now(),
				In:       true,
				MSG:      string(bmsg),
			})
			if p.Notify != nil {
				p.Notify(userID, "MSG from: "+userName, string(bmsg))
			}
			if p.UpdateHistory != nil {
				p.UpdateHistory(userID)
			}
			return
		}
	}
	SetTrans(transID, &Transfer{
		ID:       transID,
		UserID:   userID,
		DateTime: time.Now(),
		In:       true,
		Error:    e,
	})

	if p.Notify != nil {
		p.Notify(userID, "MSG from: "+userName, e.Error())
	}
	if p.UpdateHistory != nil {
		p.UpdateHistory(userID)
	}
}

func FileExist(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func CheckName(p string) string {
	ext := path.Ext(p)
	pth := p[:len(p)-len(ext)]
	num := uint64(1)
	for FileExist(p) {
		num++
		p = fmt.Sprintf("%s_(%d)%s", pth, num, ext)
	}
	return p
}

func (p *Server) GetResources(connection net.Conn) {
	userID, userName, _, transID, err := p.GetUser(connection)
	if err != nil {
		return
	}

	bint := make([]byte, 8)
	// Get bufSize
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	BufSize := BytesToInt(bint)

	// Get Total size
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	TotalBytes := BytesToInt(bint)
	// get sended bytes
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	TransBytes := BytesToInt(bint)

	// Get files
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	files_number := BytesToInt(bint)
	files := make([]*Element, files_number)
	//var num uint64
	for i := range files {
		_, err = connection.Read(bint)
		if err != nil {
			return
		}
		fileName := make([]byte, BytesToInt(bint))
		_, err = connection.Read(fileName)
		if err != nil {
			return
		}
		// File size
		_, err = connection.Read(bint)
		if err != nil {
			return
		}
		size := BytesToInt(bint)
		// File progress
		_, err = connection.Read(bint)
		if err != nil {
			return
		}
		prog := BytesToInt(bint)
		files[i] = &Element{
			Path: path.Join(p.conf.C_InboxDir, string(fileName)),
			Name: string(fileName),
			Size: size,
			Prog: prog,
		}
	}
	// Get current files
	_, err = connection.Read(bint)
	if err != nil {
		return
	}
	files_index := BytesToInt(bint)

	trans := &Transfer{
		ID:       transID,
		UserID:   userID,
		DateTime: time.Now(),
		In:       true,
		File: &FileTransfer{
			Files:      files,
			Index:      files_index,
			TransBytes: TransBytes,
			TotalBytes: TotalBytes,
			Canceled:   false,
		},
	}
	SetTrans(transID, trans)
	if p.UpdateHistory != nil {
		defer p.UpdateHistory(userID)
	}
	if p.Notify != nil {
		p.Notify(userID, "File from: "+userName, fmt.Sprintf("%d files", files_number))
	}

	// Recive files
	var f *os.File
	var dir string
	var t int
	buf := make([]byte, BufSize)
	ctl := make([]byte, 1)
	for trans.File.Index = files_index; trans.File.Index < files_number; trans.File.Index++ {
		file := trans.File.Files[trans.File.Index]

		dir = path.Dir(file.Path)
		if dir != "." {
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				trans.Error = err
				return
			}
		}

		tmp := file.Path + "_" + trans.ID + ".tmp"
		if file.Prog == 0 {
			//tmp = CheckName(tmp)
			f, err = os.Create(tmp)
			if err != nil {
				trans.Error = err
				return
			}
		} else {
			f, err = os.OpenFile(tmp, os.O_RDWR, 0777)
			if err != nil {
				trans.Error = err
				return
			}
		}
		f.Seek(int64(file.Prog), 0)

		for file.Prog < file.Size {
			// Check if source canceled
			_, err = connection.Read(ctl)
			if err != nil {
				trans.Error = err
				f.Close()
				return
			}
			if ctl[0] == CANCELED {
				trans.File.Canceled = true
				f.Close()
				return
			}

			if file.Size-file.Prog < BufSize {
				t, err = connection.Read(buf[:file.Size-file.Prog])
			} else {
				t, err = connection.Read(buf)
			}
			if err != nil {
				trans.Error = err
				f.Close()
				return
			}
			_, err = f.Write(buf[:t])
			if err != nil {
				trans.Error = err
				f.Close()
				return
			}

			// Send ctl to cancel or continue
			if trans.File.Canceled {
				_, err = connection.Write([]byte{CANCELED})
				if err != nil {
					trans.Error = err
				}
				f.Close()
				return
			}
			_, err = connection.Write([]byte{OK})
			if err != nil {
				trans.Error = err
				f.Close()
				return
			}

			file.Prog += uint64(t)
			trans.File.TransBytes += uint64(t)
			if p.UpdateHistory != nil {
				p.UpdateHistory(userID)
			}
		}
		f.Close()
		os.Rename(tmp, CheckName(file.Path))
	}
}
