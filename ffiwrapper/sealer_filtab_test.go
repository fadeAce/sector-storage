package ffiwrapper

import (
	"context"
	"fmt"
	"github.com/filecoin-project/sector-storage/ffiwrapper/basicfs"
	"github.com/filecoin-project/sector-storage/stores"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"io"
	"io/ioutil"
	"os"
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
	dir, err := ioutil.TempDir("/Users/huangchunyan/Documents/doall/filtab/", "sbtest")
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
	unpaddedSize := abi.PaddedPieceSize(2048).Unpadded()

	fmt.Println("unpadding sector size: ",unpaddedSize)

	info, err := sb.AddPiece(ctx, sectorID, nil, unpaddedSize, io.LimitReader(&PledgeReader{}, int64(unpaddedSize)))
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
	//1. 依赖AddPiece的内存结果
	//2. 依赖P1的ticket
	//3. 依赖P2的sectorCID
	//4. 依赖P1,P2的文件
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

	//verify
	//c-2做完已定要验证
	//cids, err := sb.SealPreCommit2(context.TODO(), sid, pc1o)
	//cids来自PreCommit2
	svi := abi.SealVerifyInfo{
		SectorID:              sectorID,
		SealedCID:             sectorCID.Sealed,//cids.Sealed,
		SealProof:             sb.SealProofType(),
		Proof:                 proof,
		DealIDs:               nil,
		Randomness:            ticket,
		InteractiveRandomness: seed,
		UnsealedCID:           sectorCID.Unsealed,//cids.Unsealed,
	}

	ok, err := ProofVerifier.VerifySeal(svi)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Errorf("porep proof for sector %d was invalid", sectorID.Number)
	}
	fmt.Println("verify ok!")

	//unseal

	log.Infof("[%d] Unsealing sector", sectorID.Number)

	{
		{
			p, done, err := sp.AcquireSector(context.TODO(), sectorID, stores.FTUnsealed, stores.FTNone, true)
			if err != nil {
				t.Errorf("acquire unsealed sector for removing: %w", err)
			}
			done()

			fmt.Println("Remove unsealed files:", p.Unsealed)

			if err := os.Remove(p.Unsealed); err != nil {
				t.Errorf("removing unsealed sector: %w", err)
			}
		}

		//这一步做了啥？
		err := sb.UnsealPiece(context.TODO(), sectorID, 0, unpaddedSize, ticket, sectorCID.Unsealed)
		if err != nil {
			t.Error(err)
		}
	}

	fmt.Println("Finish unsealed")
}
