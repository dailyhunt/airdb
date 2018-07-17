package commonUtils

const OneMb = 1048576

func StrToBytes(s string) []byte {
	return []byte(s)
}

func MbToBytes(mb int) int {
	return mb * OneMb
}
