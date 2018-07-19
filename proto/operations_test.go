package proto

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestPut(t *testing.T) {
	put := &Put{
		Key:   []byte("k1"),
		Col:   []byte("ig"),
		Value: []byte("ig_x_256"),
		Epoch: uint64(1),
	}

	dAtA, _ := put.Marshal()
	assert.NotNil(t, put.String())
	assert.Equal(t, 20, len(dAtA), "number of bytes should be same")
}

func TestPutBatch10(t *testing.T) {
	batch := PutBatch{}
	fmt.Println(unsafe.Sizeof(batch))

	var puts []*Put

	for i := 0; i < 10; i++ {
		p := &Put{
			Key:   []byte("k1"),
			Col:   []byte("ig"),
			Value: []byte("ig_x_256"),
			Epoch: uint64(1),
		}
		puts = append(puts, p)
	}

	batch.Puts = puts

	dAtA, _ := batch.Marshal()
	assert.NotNil(t, batch.String())
	assert.Equal(t, 220, len(dAtA), "number of bytes should be same")
}

func TestPutBatch2(t *testing.T) {
	batch := PutBatch{}
	fmt.Println(unsafe.Sizeof(batch))

	var puts []*Put

	for i := 0; i < 2; i++ {
		p := &Put{
			Key:   []byte("k1"),
			Col:   []byte("ig"),
			Value: []byte("ig_x_256"),
			Epoch: uint64(1),
		}
		puts = append(puts, p)
	}

	batch.Puts = puts

	dAtA, _ := batch.Marshal()
	assert.NotNil(t, batch.String())
	// put instances (20 * 2)  + array overhead (2*2)
	assert.Equal(t, 44, len(dAtA), "number of bytes should be same")
}
