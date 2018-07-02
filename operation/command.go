package operation

import (
	"github.com/gin-gonic/gin/json"
	"time"
	"unsafe"
)

const PUT_SIZE = uint64(unsafe.Sizeof(&Put{}))

type Op interface {
	Name() string
	String() string
	Size() uint64
}

type Put struct {
	K string
	V string
	T time.Time
}

func (p *Put) Size() uint64 {
	return PUT_SIZE + uint64(len(p.K)+len(p.V))
}

func (p *Put) Name() string {
	return "PUT"
}

func (p *Put) String() string {
	bytes, _ := json.Marshal(p)
	return string(bytes)
}

type Get struct {
	K string
	V string
}

func (*Get) Name() string {
	return "GET"
}
