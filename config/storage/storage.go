package storage

import (
	"os"
	"os/user"
	"path"
	"sort"

	"gioui.org/layout"
	"gioui.org/widget"
)

type Element struct {
	IsDir bool
	Name  string
	Path  string

	Anim      float32
	Dim       layout.Dimensions
	Clickable widget.Clickable
	Selected  widget.Bool
}

func Explore(p string) ([]*Element, error) {
	if p == "" {
		return GetStorages(), nil
	}
	elements := []*Element{}
	res, e := os.ReadDir(p)
	if e != nil {
		return elements, e
	}
	for _, r := range res {
		elements = append(elements, &Element{
			IsDir: r.IsDir(),
			Name:  r.Name(),
			Path:  path.Join(p, r.Name()),
			Anim:  0,
		})
	}
	sort.Slice(elements, func(i, j int) bool {
		b := elements[i].Name < elements[j].Name
		if elements[i].IsDir && !elements[j].IsDir {
			b = true
		} else if elements[j].IsDir && !elements[i].IsDir {
			b = false
		}
		return b
	})
	return elements, nil
}

func GetStorages() []*Element {
	elements := []*Element{}
	places := getStorages()
	for _, r := range places {
		_, n := path.Split(r)
		elements = append(elements, &Element{
			IsDir: true,
			Name:  n,
			Path:  r,
			Anim:  0,
		})
	}
	return elements
}

func getUsername() string {
	user, err := user.Current()
	if err == nil {
		return user.Username
	}
	return ""
}

func checkStorage(p string) bool {
	//_, e := os.ReadDir(p)
	//return e == nil
	inf, err := os.Stat(p)
	return err == nil && inf.IsDir()
}
