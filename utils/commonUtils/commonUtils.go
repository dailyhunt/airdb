package commonUtils

import (
	"encoding/binary"
	"math"
)

const OneMb = 1048576

func StrToBytes(s string) []byte {
	return []byte(s)
}

func MbToBytes(mb int) int {
	return mb * OneMb
}

func KeyWithTs(key []byte, ts uint64) []byte {
	out := make([]byte, len(key)+8)
	copy(out, key)
	binary.BigEndian.PutUint64(out[len(key):], math.MaxUint64-ts)
	return out
}
