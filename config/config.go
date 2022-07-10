package config

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget/material"
)

const (
	Port     = 9182
	ConfFile = "config"
)

const (
	ICSubNetworks = '\uf0e8'
	ICConnections = '\uf819'
	ICInbox       = '\uf7fa'
	ICConfig      = '\ue615'

	ICNewSubnet = '\uf844'
	ICDelete    = '\uf658'

	ICScan     = '\uf43c'
	ICScanStop = '\uf28d'
	ICSend     = '\uf1d8'
	ICClose    = '\uf658'

	ICAndroid = '\uf17b'
	ICApple   = '\uf179'
	ICWindows = '\uf17a'
	ICLinux   = '\uf17c'
	ICUnknow  = '\uf29c'

	ICStorages = '\uf0a0'
	ICStorage  = '\uf0a0'
	ICOpenDir  = '\uf115'
	ICDirUp    = '\uf63c'
	ICDir      = '\uf114'
	ICFile     = '\uf016'
	ICMSG      = '\uf430'

	ICSave  = '\uf0c7'
	ICReset = '\uf0e2'
)

type Config struct {
	th *material.Theme

	C_Name               string
	C_InboxDir           string
	C_Connections        uint64
	C_ConnectionsTimeout uint64
	C_BufSize            uint64

	ScreenColor color.NRGBA
	Shadow      color.NRGBA

	BGColor color.NRGBA
	FGColor color.NRGBA

	BGPrimaryColor color.NRGBA
	FGPrimaryColor color.NRGBA

	DangerColor     color.NRGBA
	AndroidBarColor color.NRGBA

	RecivedColor color.NRGBA
	SendedColor  color.NRGBA

	openDialog  func(func(*material.Theme, layout.Context, *app.Window, *Config) layout.Dimensions)
	closeDialog func()
}

func NewConfig(th *material.Theme) *Config {
	conf := &Config{
		th: th,
	}
	if !conf.Load() {
		conf.Reset()
	}
	conf.UpdateColors()
	return conf
}

func (p *Config) Reset() {
	p.C_Name = p.GetName()
	p.C_InboxDir = path.Join(p.AppDir(), "files")
	p.C_Connections = 20
	p.C_ConnectionsTimeout = 500
	p.C_BufSize = 1024
	os.MkdirAll(p.C_InboxDir, 0777)

	p.ScreenColor = color.NRGBA{230, 230, 230, 255}
	p.Shadow = color.NRGBA{0, 0, 0, 100}

	p.BGColor = color.NRGBA{255, 255, 255, 255}
	p.FGColor = color.NRGBA{100, 100, 100, 255}

	p.BGPrimaryColor = color.NRGBA{0, 100, 255, 255}
	p.FGPrimaryColor = color.NRGBA{255, 255, 255, 255}

	p.DangerColor = color.NRGBA{255, 0, 0, 255}
	p.AndroidBarColor = color.NRGBA{0, 50, 255, 255}

	p.RecivedColor = color.NRGBA{255, 255, 255, 255}
	p.SendedColor = color.NRGBA{200, 255, 255, 255}

	p.UpdateColors()
}

func (p *Config) Load() bool {
	filePath := path.Join(p.AppDir(), ConfFile)
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}
	return json.Unmarshal(buf, p) == nil
}

func (p *Config) Save() error {
	buf, err := json.Marshal(p)
	if err != nil {
		return err
	}
	filePath := path.Join(p.AppDir(), ConfFile)
	return ioutil.WriteFile(filePath, buf, 0777)
}

func (p *Config) Name() string {
	return p.C_Name
}

func (p *Config) Inbox() string {
	return p.C_InboxDir
}

func (p *Config) Connections() uint64 {
	return p.C_Connections
}

func (p *Config) Timeout() uint64 {
	return p.C_ConnectionsTimeout
}

func (p *Config) BufSize() uint64 {
	return p.C_BufSize
}

func (p *Config) SetName(n string) error {
	p.C_Name = n
	return p.Save()
}

func (p *Config) SetInbox(d string) error {
	p.C_InboxDir = d
	return p.Save()
}

func (p *Config) SetConnections(n uint64) error {
	p.C_Connections = n
	return p.Save()
}

func (p *Config) SetTimeout(n uint64) error {
	p.C_ConnectionsTimeout = n
	return p.Save()
}

func (p *Config) SetBufSize(n uint64) error {
	p.C_BufSize = n
	return p.Save()
}

func (p *Config) OS() string {
	return runtime.GOOS
}

func (p *Config) AppDir() string {
	dir, err := app.DataDir()
	if err != nil {
		dir = ""
	} else {
		dir = path.Join(dir, "JG_Sender")
	}
	return dir
}

func (p *Config) GetName() string {
	var name string
	user, err := user.Current()
	if err == nil {
		name = user.Username
	} else {
		name, err = os.Hostname()
		if err != nil {
			name = "MyName"
		}
	}
	return name
}

func (p *Config) UpdateColors() {
	p.th.Bg = p.BGColor
	p.th.Fg = p.FGColor
	p.th.ContrastBg = p.BGPrimaryColor
	p.th.ContrastFg = p.FGPrimaryColor
}

func (p *Config) OpenDialog(d func(*material.Theme, layout.Context, *app.Window, *Config) layout.Dimensions) {
	if p.openDialog != nil {
		p.openDialog(d)
	}
}

func (p *Config) CloseDialog() {
	if p.closeDialog != nil {
		p.closeDialog()
	}
}

func (p *Config) SetDialogOpener(opener func(func(*material.Theme, layout.Context, *app.Window, *Config) layout.Dimensions)) {
	p.openDialog = opener
}

func (p *Config) SetDialogCloser(closer func()) {
	p.closeDialog = closer
}
