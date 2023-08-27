package mtreeimpl

import (
	"bytes"
	"crypto/sha256"
	"hash"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

// VerifyMerklePath is a public function that helps to verify the merkle Path and the content
// to check if it match with the merkle root input it provided.
// if hashStrategy in the parameter is nil, then it will use default hash sha256
func VerifyMerklePath(c types.KVScontent, mpath []types.MerklePath, mroot []byte, hashStrategy func() hash.Hash) (bool, error) {
	if hashStrategy == nil {
		hashStrategy = sha256.New
	}
	calMroot, err := c.CalculateHash()
	if err != nil {
		return false, err
	}
	for _, p := range mpath {
		h := hashStrategy()
		if p.Pos {
			_, err = h.Write(append(calMroot, p.Path...))
		} else {
			_, err = h.Write(append(p.Path, calMroot...))
		}
		if err != nil {
			return false, err
		}
		calMroot = h.Sum(nil)
	}
	return bytes.Equal(mroot, calMroot), nil
}
