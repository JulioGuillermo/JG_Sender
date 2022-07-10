package storage

func getStorages() []string {
	storages := []string{}
	l := 'a'
	for l <= 'z' {
		if checkStorage(string(l) + ":") {
			storages = append(storages, string(l)+":")
		}
		l++
	}
	return storages
}
