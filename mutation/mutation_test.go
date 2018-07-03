package mutation

import (
	ut "github.com/dailyhunt/airdb/utils/commonUtils"
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestEncodeSingleValue(t *testing.T) {
	mut := prepareMutationObject("1234", "fam1", "col1", "airdb", PUT)

	want := calculateBuffSize(mut)
	got := len(Encode(mut))
	assert.Equal(t, got, want, "Encoded byte size is as expected")
}

func TestDecodeSingleValue(t *testing.T) {
	key := "1234"
	value := "airdb"
	fam := "fam1"
	col := "col1"
	mType := PUT
	mut := prepareMutationObject(key, fam, col, value, mType)
	encoded := Encode(mut)

	dm, err := Decode(encoded)

	if err != nil {
		t.Error("Error while decoding byte to mutation")
	}

	assert.Equal(t, string(dm.Key), key)
	assert.Equal(t, string(dm.Value), value)
	assert.Equal(t, string(dm.Family), fam)
	assert.Equal(t, string(dm.Col), col)
	assert.Equal(t, dm.MutationType, mType)
	assert.Equal(t, dm.Timestamp > 0, true)

}

// Encoded Byte
func prepareMutationObject(k, cf, col, v string, mt Type) *Mutation {
	return &Mutation{
		Key:          ut.StrToBytes(k),
		Family:       ut.StrToBytes(cf),
		Col:          ut.StrToBytes(col),
		Value:        ut.StrToBytes(v),
		Timestamp:    uint64(time.Now().Unix()),
		MutationType: mt,
	}
}
