//go:build (linux && !android) || freebsd || openbsd
// +build linux,!android freebsd openbsd

package storage

import (
	"os"
	"path"
)

func getStorages() []string {
	storages := []string{}

	username := getUsername()
	if username != "" {
		if checkStorage("/") {
			storages = append(storages, "/")
		}

		home := path.Join("/home", username)
		if checkStorage(home) {
			storages = append(storages, home)
		}

		media := path.Join("/media", username)
		if checkStorage(media) {
			res, err := os.ReadDir(media)
			if err == nil {
				for _, r := range res {
					if r.IsDir() {
						storages = append(storages, path.Join(media, r.Name()))
					}
				}
			}
		}
		media = path.Join("/run/media", username)
		if checkStorage(media) {
			res, err := os.ReadDir(media)
			if err == nil {
				for _, r := range res {
					if r.IsDir() {
						storages = append(storages, path.Join(media, r.Name()))
					}
				}
			}
		}
	}

	return storages
}
