package connection

import (
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/julioguillermo/jg_sender/config"
)

func (p *Server) SendUser(connection net.Conn, transID string) error {
	userID := p.conf.UUID
	userName := p.conf.C_Name
	userOS := p.conf.OS()

	// User ID
	_, e := connection.Write(IntToBytes(uint64(len(userID))))
	if e != nil {
		return e
	}
	_, e = connection.Write([]byte(userID))
	if e != nil {
		return e
	}

	// User Name
	_, e = connection.Write(IntToBytes(uint64(len(userName))))
	if e != nil {
		return e
	}
	_, e = connection.Write([]byte(userName))
	if e != nil {
		return e
	}

	// User OS
	_, e = connection.Write(IntToBytes(uint64(len(userOS))))
	if e != nil {
		return e
	}
	_, e = connection.Write([]byte(userOS))
	if e != nil {
		return e
	}

	// Trans ID
	_, e = connection.Write(IntToBytes(uint64(len(transID))))
	if e != nil {
		return e
	}
	_, e = connection.Write([]byte(transID))
	if e != nil {
		return e
	}
	return nil
}

func (p *Server) SendMSG(userID string, msg string) {
	device := GetDevice(userID)
	transID := uuid.NewString()

	trans := &Transfer{
		ID:       transID,
		UserID:   userID,
		DateTime: time.Now(),
		In:       false,
		MSG:      msg,
	}
	SetTrans(transID, trans)
	if p.UpdateHistory != nil {
		p.UpdateHistory(userID)
		defer p.UpdateHistory(userID)
	}

	addrPort := netip.AddrPortFrom(*device.Addr, uint16(config.Port))
	connection, e := net.Dial("tcp", addrPort.String())
	if e != nil {
		trans.Error = e
		return
	}

	_, e = connection.Write([]byte{MSG})
	if e != nil {
		trans.Error = e
		return
	}

	e = p.SendUser(connection, transID)
	if e != nil {
		trans.Error = e
		return
	}

	_, e = connection.Write(IntToBytes(uint64(len(msg))))
	if e != nil {
		trans.Error = e
		return
	}
	_, e = connection.Write([]byte(msg))
	if e != nil {
		trans.Error = e
		return
	}
}

func (p *Server) SendResources(userID string, resources []string) {
	// Getting total size and files
	tsize := uint64(0)
	files := []*Element{}
	var (
		inf        fs.FileInfo
		element    *Element
		subelement *Element
		res        []fs.DirEntry
		e          error
	)
	for _, r := range resources { // Read all resources
		inf, e = os.Stat(r)
		if e == nil {
			if inf.IsDir() { // if resource is a dir
				element = &Element{
					Path: r,
					Name: path.Base(r),
				}
				fifo := []*Element{element}
				for len(fifo) > 0 {
					element = fifo[0]
					fifo = fifo[1:]
					res, e = os.ReadDir(element.Path) // explore dir
					if e == nil {
						for _, r := range res { // get all element on dir
							subelement = &Element{
								Path: path.Join(element.Path, r.Name()),
								Name: path.Join(element.Name, r.Name()),
							}
							if r.IsDir() { // if element is a subdir
								fifo = append(fifo, subelement) // explore too
							} else {
								inf, e = os.Stat(subelement.Path) // if is a file
								if e == nil {
									subelement.Size = uint64(inf.Size())
									files = append(files, subelement) // add it and it's size
									tsize += subelement.Size
								}
							}
						}
					}
				}
			} else { // if resource is a file
				element = &Element{
					Path: r,
					Name: path.Base(r),
					Size: uint64(inf.Size()),
				}
				files = append(files, element) // Add to files
				tsize += element.Size          // add it's size
			}
		}
	}

	transID := uuid.NewString()

	trans := &Transfer{
		ID:       transID,
		UserID:   userID,
		DateTime: time.Now(),
		In:       false,
		File: &FileTransfer{
			Files:      files,
			TotalBytes: tsize,
		},
	}
	SetTrans(transID, trans)
	if p.UpdateHistory != nil {
		p.UpdateHistory(userID)
		defer p.UpdateHistory(userID)
	}

	p.SendTrans(userID, trans)
}

func (p *Server) ContinueTrans(userID string, trans *Transfer) {
	trans.Error = nil
	trans.File.Canceled = false
	p.UpdateHistory(userID)
	if trans.In {

	} else {
		p.SendTrans(userID, trans)
	}
}

func (p *Server) SendTrans(userID string, trans *Transfer) {
	if p.UpdateHistory != nil {
		defer p.UpdateHistory(userID)
	}

	device := GetDevice(userID)

	// Connecting
	addrPort := netip.AddrPortFrom(*device.Addr, uint16(config.Port))
	connection, e := net.Dial("tcp", addrPort.String())
	if e != nil {
		trans.Error = e
		return
	}
	defer connection.Close()

	// CTL MSG: RESOURCES
	_, e = connection.Write([]byte{RESOURCES})
	if e != nil {
		trans.Error = e
		return
	}
	// Send user info and trans id
	e = p.SendUser(connection, trans.ID)
	if e != nil {
		trans.Error = e
		return
	}

	// Buf size
	_, e = connection.Write(IntToBytes(p.conf.BufSize()))
	if e != nil {
		trans.Error = e
		return
	}

	// Total size and total progress
	_, e = connection.Write(IntToBytes(trans.File.TotalBytes))
	if e != nil {
		trans.Error = e
		return
	}
	_, e = connection.Write(IntToBytes(trans.File.TransBytes))
	if e != nil {
		trans.Error = e
		return
	}

	// files
	_, e = connection.Write(IntToBytes(uint64(len(trans.File.Files))))
	if e != nil {
		trans.Error = e
		return
	}
	for _, f := range trans.File.Files {
		// Send file name
		_, e = connection.Write(IntToBytes(uint64(len(f.Name))))
		if e != nil {
			trans.Error = e
			return
		}
		_, e = connection.Write([]byte(f.Name))
		if e != nil {
			trans.Error = e
			return
		}
		// File size
		_, e = connection.Write(IntToBytes(f.Size))
		if e != nil {
			trans.Error = e
			return
		}
		// File progress
		_, e = connection.Write(IntToBytes(f.Prog))
		if e != nil {
			trans.Error = e
			return
		}
	}

	// current file
	_, e = connection.Write(IntToBytes(trans.File.Index))
	if e != nil {
		trans.Error = e
		return
	}

	buf := make([]byte, p.conf.BufSize())
	ctl := make([]byte, 1)
	for trans.File.Index < uint64(len(trans.File.Files)) {
		file := trans.File.Files[trans.File.Index]
		fr, e := os.Open(file.Path)
		if e != nil {
			trans.Error = e
			return
		}
		_, e = fr.Seek(int64(file.Prog), 0)
		if e != nil {
			trans.Error = e
			return
		}

		for file.Prog < file.Size {
			// Send ctl to cancel or continue
			if trans.File.Canceled {
				_, e = connection.Write([]byte{CANCELED})
				if e != nil {
					trans.Error = e
				}
				return
			}
			_, e = connection.Write([]byte{OK})
			if e != nil {
				trans.Error = e
				return
			}

			t, e := fr.Read(buf)
			if e != nil {
				trans.Error = e
				return
			}
			_, e = connection.Write(buf[:t])
			if e != nil {
				trans.Error = e
				return
			}

			// Check if destiny canceled
			_, e = connection.Read(ctl)
			if e != nil {
				trans.Error = e
				return
			}
			if ctl[0] != OK {
				trans.File.Canceled = true
				return
			}

			file.Prog += uint64(t)
			trans.File.TransBytes += uint64(t)
			if p.UpdateHistory != nil {
				p.UpdateHistory(userID)
			}
		}
		trans.File.Index++
	}
}
