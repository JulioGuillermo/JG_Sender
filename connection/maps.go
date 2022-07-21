package connection

import (
	"net/netip"
	"time"
)

type Element struct {
	Path string
	Name string
	Size uint64
	Prog uint64
}

type FileTransfer struct {
	Files      []*Element
	Index      uint64
	TransBytes uint64
	TotalBytes uint64
	Canceled   bool
}

type Transfer struct {
	ID       string
	UserID   string
	DateTime time.Time
	In       bool
	View     bool
	MSG      string
	Error    error
	File     *FileTransfer
}

type Device struct {
	ID     string
	Addr   *netip.Addr
	Name   string
	OS     string
	Not    uint64
	Online bool
}

var History = []*Transfer{}
var Devices = []*Device{}

// History ######################################
func GetUserHistory(userid string) []*Transfer {
	his := []*Transfer{}
	for _, t := range History {
		if t.UserID == userid {
			his = append(his, t)
		}
	}
	return his
}

func UserView(userid string) {
	for _, t := range History {
		if t.UserID == userid {
			t.View = true
		}
	}
}

func FindTrans(id string) int {
	for i, d := range History {
		if d.ID == id {
			return i
		}
	}
	return -1
}

func GetTrans(id string) *Transfer {
	index := FindTrans(id)
	if index == -1 {
		return nil
	}
	return History[index]
}

func SetTrans(id string, d *Transfer) {
	index := FindTrans(id)
	if index == -1 {
		History = append(History, d)
		return
	}
	History[index] = d
}

// Device ######################################
func FindDevice(id string) int {
	for i, d := range Devices {
		if d.ID == id {
			return i
		}
	}
	return -1
}

func GetDevice(id string) *Device {
	index := FindDevice(id)
	if index == -1 {
		return nil
	}
	return Devices[index]
}

func SetDevice(id string, d *Device) {
	index := FindDevice(id)
	if index != -1 {
		d.Not = Devices[index].Not
		Devices = append(Devices[:index], Devices[index+1:]...)
	}
	d.Online = true
	Devices = append([]*Device{d}, Devices...)
}

func InvalidateDevices() {
	for _, d := range Devices {
		d.Online = false
	}
}
