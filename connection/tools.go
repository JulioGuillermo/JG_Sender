package connection

const (
	NAME = byte(iota)
	MSG
	RESOURCES
)

func IntToBytes(num uint64) []byte {
	bs := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bs[i] = byte(num >> (i * 8))
	}
	return bs
}

func BytesToInt(bs []byte) uint64 {
	res := uint64(0)
	for i := 0; i < len(bs); i++ {
		res += uint64(bs[i]) << (i * 8)
	}
	return res
}
