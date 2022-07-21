package connection

const (
	NAME = byte(iota)
	MSG
	RESOURCES

	EXPLORE
	GET
	EXEC_CMD

	DIR
	FILE

	ERROR

	CANCELED
	OK
)

var CTL = []byte{0, 2, 0, 8, 2, 0, 0, 0}

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

func CheckCTL(ctl []byte) bool {
	if len(ctl) != len(CTL) {
		return false
	}
	for i := range CTL {
		if CTL[i] != ctl[i] {
			return false
		}
	}
	return true
}
