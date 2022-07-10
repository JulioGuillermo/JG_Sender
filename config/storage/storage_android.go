package storage

import (
	"os"
	"path"
)

func getStorages() []string {
	storages := []string{}
	if checkStorage("/sdcard") {
		storages = append(storages, "/sdcard")
	}
	dirs, e := os.ReadDir("/storage")
	if e == nil {
		for _, d := range dirs {
			s := path.Join("/storage", d.Name())
			if checkStorage(s) {
				storages = append(storages, s)
			}
		}
	}
	return storages
}
