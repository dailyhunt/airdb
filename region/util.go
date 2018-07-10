package region

import (
	"encoding/binary"
	"github.com/satori/go.uuid"
)

func CreateRegionId() uint64 {
	b := uuid.NewV4()
	return binary.BigEndian.Uint64(b[:16])
}
