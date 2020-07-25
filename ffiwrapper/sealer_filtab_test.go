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

func TestSealer_integration(t *testing.T) {
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

	ctx := context.TODO()
	// add piece
	info, err := sb.AddPiece(ctx, sectorID, nil, 2032, io.LimitReader(&PledgeReader{}, int64(2032)))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(info)

	// pre-1
	ticket := abi.SealRandomness{9, 9, 9, 9, 9, 9, 9, 9}
	pre1out, err := sb.SealPreCommit1(ctx, sectorID, ticket, []abi.PieceInfo{info})
	if err != nil {
		t.Error(err)
	}
	fmt.Println(pre1out)

	// pre-2
	sectorCID, err := sb.SealPreCommit2(ctx, sectorID, pre1out)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(sectorCID)

	// c-1
	seed := abi.InteractiveSealRandomness{0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 45, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 9}
	c1out, err := sb.SealCommit1(ctx, sectorID, ticket, seed, []abi.PieceInfo{info}, sectorCID)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(c1out)

	// c-2
	proof, err := sb.SealCommit2(ctx, sectorID, c1out)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(proof)

}
