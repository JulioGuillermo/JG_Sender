package connection

import (
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path"

	"github.com/julioguillermo/jg_sender/config"
)

type Element struct {
	path string
	name string
	size uint64
}

type Sender struct {
	sending    bool
	connection net.Conn
}

func NewSender() *Sender {
	return &Sender{
		sending:    false,
		connection: nil,
	}
}

func (p *Sender) Close() {
	p.sending = false
	if p.connection != nil {
		p.connection.Close()
		p.connection = nil
	}
}

/**
Send:
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

func (p *Sender) SendResources(conf *config.Config, addr *netip.Addr, resources []string, onError func(error), onProgress func(uint64, uint64, uint64)) {
	defer p.Close()

	var err error
	// Open connection
	addrPort := netip.AddrPortFrom(*addr, uint16(config.Port))
	p.connection, err = net.Dial("tcp", addrPort.String())
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}

	p.sendResources(conf, resources, onError, onProgress)
}

func (p *Sender) sendResources(conf *config.Config, resources []string, onError func(error), onProgress func(uint64, uint64, uint64)) {
	var err error

	if onError == nil {
		onError = func(error) {}
	}
	if onProgress == nil {
		onProgress = func(uint64, uint64, uint64) {}
	}

	// Getting total size and files
	tsize := uint64(0)
	files := []*Element{}
	var (
		inf        fs.FileInfo
		element    *Element
		subelement *Element
		res        []fs.DirEntry
	)
	for _, r := range resources { // Read all resources
		inf, err = os.Stat(r)
		if err == nil {
			if inf.IsDir() { // if resource is a dir
				element = &Element{
					path: r,
					name: path.Base(r),
				}
				fifo := []*Element{element}
				for len(fifo) > 0 {
					element = fifo[0]
					fifo = fifo[1:]
					res, err = os.ReadDir(element.path) // explore dir
					if err == nil {
						for _, r := range res { // get all element on dir
							subelement = &Element{
								path: path.Join(element.path, r.Name()),
								name: path.Join(element.name, r.Name()),
							}
							if r.IsDir() { // if element is a subdir
								fifo = append(fifo, subelement) // explore too
							} else {
								inf, err = os.Stat(subelement.path) // if is a file
								if err == nil {
									subelement.size = uint64(inf.Size())
									files = append(files, subelement) // add it and it's size
									tsize += subelement.size
								}
							}
						}
					}
				}
			} else { // if resource is a file
				element = &Element{
					path: r,
					name: path.Base(r),
					size: uint64(inf.Size()),
				}
				files = append(files, element) // Add to files
				tsize += element.size          // add it's size
			}
		}
	}

	// CTL MSG: RESOURCES
	_, err = p.connection.Write([]byte{RESOURCES})
	if err != nil {
		onError(err)
		return
	}

	// USER Name
	_, err = p.connection.Write(IntToBytes(uint64(len(conf.C_Name))))
	if err != nil {
		onError(err)
		return
	}
	_, err = p.connection.Write([]byte(conf.C_Name))
	if err != nil {
		onError(err)
		return
	}

	// BufSize, total size and number of files
	_, err = p.connection.Write(IntToBytes(conf.BufSize()))
	if err != nil {
		onError(err)
		return
	}
	_, err = p.connection.Write(IntToBytes(tsize))
	if err != nil {
		onError(err)
		return
	}
	_, err = p.connection.Write(IntToBytes(uint64(len(files))))
	if err != nil {
		onError(err)
		return
	}

	// Files names
	for _, f := range files {
		_, err = p.connection.Write(IntToBytes(uint64(len(f.name))))
		if err != nil {
			onError(err)
			return
		}
		_, err = p.connection.Write([]byte(f.name))
		if err != nil {
			onError(err)
			return
		}
	}
	// Files
	buf := make([]byte, conf.BufSize())
	t_pro := uint64(0)
	var pro uint64
	var t int
	var file *os.File

	for i, f := range files {
		_, err = p.connection.Write(IntToBytes(f.size))
		if err != nil {
			onError(err)
			return
		}
		pro = 0
		file, err = os.Open(f.path)
		if err != nil {
			onError(err)
			return
		}
		for pro < f.size {
			t, err = file.Read(buf)
			if err != nil {
				onError(err)
				return
			}
			_, err = p.connection.Write(buf[:t])
			if err != nil {
				onError(err)
				return
			}
			pro += uint64(t)
			t_pro += uint64(t)
			onProgress(uint64(i), t_pro, tsize)
		}
	}
	onProgress(uint64(len(files)), tsize, tsize)
}
