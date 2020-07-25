package ffiwrapper

import (
	"context"
	"fmt"
	"github.com/filecoin-project/sector-storage/ffiwrapper/basicfs"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"io"
	"io/ioutil"
	"testing"
)

// TODO: extract this to someplace where it can be shared with lotus
type PledgeReader struct{}

func (PledgeReader) Read(out []byte) (int, error) {
	for i := range out {
		out[i] = 0
	}
	return len(out), nil
}

func TestSealer_AddPiece(t *testing.T) {

	dir, err := ioutil.TempDir("", "sbtest")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		SealProofType: sealProofType,
	}
	sp := &basicfs.Provider{
		Root: dir,
	}

	sb, err := New(sp, cfg)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	sectorID := abi.SectorID{}
	info, err := sb.AddPiece(context.TODO(), sectorID, nil, 2032, io.LimitReader(&PledgeReader{}, int64(2032)))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(info)
}
