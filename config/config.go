package config

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/google/uuid"
)

const (
	FPS             = 25
	updatePerSecond = time.Second / FPS
)

const (
	Port     = 9182
	ConfFile = "config"
)

const (
	ICSubNetworks = '\uf0e8'
	ICConnections = '\uf819'
	ICConfig      = '\ue615'

	ICAndroid = '\uf17b'
	ICApple   = '\uf179'
	ICWindows = '\uf17a'
	ICLinux   = '\uf17c'
	ICUnknow  = '\uf29c'

	ICOpenFile = '\uf89b'
	ICOpenDir  = '\uf115'
	ICDirUp    = '\uf63c'

	ICStorages = '\uf0a0'
	ICStorage  = '\uf0a0'
	ICDir      = '\uf114'
	ICFile     = '\uf016'

	ICAdd    = '\uf4a7'
	ICClose  = '\uf467'
	ICUpdate = '\uf46a'
	ICReset  = '\uf0e2'
	ICBack   = '\uf4a8'
	ICSend   = '\uf1d8'
	ICOK     = '\uf62b'

	ICOnline  = '\uf836'
	ICOffline = '\uf837'
)

type Config struct {
	th *material.Theme

	UUID string

	C_Name               string
	C_InboxDir           string
	C_Connections        uint64
	C_ConnectionsTimeout uint64
	C_BufSize            uint64
	C_AnimTime           uint64

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

	old time.Time
}

func NewConfig(th *material.Theme) *Config {
	conf := &Config{
		th: th,
	}
	if !conf.Load() {
		conf.Reset()
	}
	if conf.UUID == "" {
		conf.UUID = uuid.NewString()
		conf.Save()
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
	p.C_AnimTime = 300
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
	c := json.Unmarshal(buf, p) == nil
	return c
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

func (p *Config) AnimTime() time.Duration {
	if p.C_AnimTime == 0 {
		return time.Millisecond
	}
	return time.Duration(p.C_AnimTime) * time.Millisecond
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

func (p *Config) SetAnimTime(ms uint64) error {
	p.C_AnimTime = ms
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

func (p *Config) AnimSpeed(gtx layout.Context) float32 {
	if p.C_AnimTime == 0 {
		return 1
	}
	t := 1000 / float32(p.C_AnimTime) / float32(FPS)
	if t > 1 {
		return 1
	}
	return t
}

func (p *Config) Time(gtx layout.Context) time.Duration {
	d := gtx.Now.Sub(p.old)
	if d < 0 {
		d = -d
	}
	p.old = time.Now()
	return updatePerSecond - d
}
