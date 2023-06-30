package mtreecomp

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/mtreeimpl"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/stretchr/testify/require"
)

func Setup(t *testing.T) *MerkleTreeComponent {
	// Noted that the merkle tree component will be only create once in all the test.
	mtci, err := NewMerkleTreeComponent()
	require.NoError(t, err)
	fmt.Println("finish create component")

	batch := statedb.NewUpdateBatch()
	batch.Put("ns1", "key1", []byte("value1"), version.NewHeight(1, 1))
	savePoint := version.NewHeight(2, 22)

	fmt.Println("finish create batch: ", batch, ", savepoint:", savePoint)

	err = mtci.ApplyUpdates(batch, savePoint)
	require.NoError(t, err)
	return mtci
}

func TestApplyUpdates(t *testing.T) {
	mtci := Setup(t)

	ns := "ns1"
	c := types.KVScontent{
		Key:   "key1",
		Value: []byte("value1"),
	}

	valid, err := mtci.VerifyContent(ns, c)
	require.NoError(t, err)
	require.True(t, valid)

	newBatch := statedb.NewUpdateBatch()
	newBatch.Put("ns1", "key1", []byte("updateValue1"), version.NewHeight(1, 2))
	savePoint := version.NewHeight(2, 23)
	err = mtci.ApplyUpdates(newBatch, savePoint)
	require.NoError(t, err)

	newC := types.KVScontent{
		Key:   "key1",
		Value: []byte("updateValue1"),
	}
	valid, err = mtci.VerifyContent(ns, newC)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestGetMerkleRoot(t *testing.T) {
	mtci := Setup(t)

	// test non exist namespace
	nsNotExist := "nsX"
	mroot, err := mtci.GetMerkleRoot(nsNotExist)
	require.NoError(t, err)
	require.Nil(t, nil)

	// test exist namespace
	nsExist := "ns1"
	mroot, err = mtci.GetMerkleRoot(nsExist)
	require.NoError(t, err)
	fmt.Println("Merkle root:", mroot)
	require.NotNil(t, mroot)
}

func TestGetMerklePath(t *testing.T) {
	mtci := Setup(t)

	ns := "ns1"
	c := types.KVScontent{
		Key:   "key1",
		Value: []byte("value1"),
	}
	mroot, err := mtci.GetMerkleRoot(ns)
	require.NoError(t, err)

	mpath, err := mtci.GetMerklePath(ns, c)
	require.NoError(t, err)
	fmt.Println("MerklePath:", mpath)

	res, err := mtreeimpl.VerifyMerklePath(c, mpath, mroot, sha256.New)
	require.NoError(t, err)
	require.True(t, res)
}
