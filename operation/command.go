package operation

import (
	"github.com/gin-gonic/gin/json"
	"time"
)

type Op interface {
	Name() string
	String() string
}

type Put struct {
	K string
	V string
	T time.Time
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
