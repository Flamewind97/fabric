package types

import (
	"crypto/sha256"
)

type KVScontent struct {
	Key   string
	Value []byte
}

func (k *KVScontent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(append([]byte(k.Key+"|"), k.Value...)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

type MerklePath struct {
	Path []byte
	Pos  bool
}

type MerkleTree interface {
	// GetMerkleRoot returns the unverified Merkle Root (hash of the root node) of the tree.
	GetMerkleRoot() []byte
	// GetMerklePath: Get Merkle Path and indexes(left leaf or right leaf)
	GetMerklePath(content KVScontent) ([]MerklePath, error)
	// VerifyContent indicates whether a given content is in the tree and the hashes are valid for that content.
	// Returns true if the expected Merkle Root is equivalent to the Merkle root calculated on the critical Path
	// for a given content. Returns true if valid and false otherwise.
	VerifyContent(content KVScontent) (bool, error)
	// Add add a new content into the merkle tree
	Add(content KVScontent) error
	// Delete delete the content from the merkle tree
	Delete(content KVScontent) error
	// Update update the old content in the merkle tree
	Update(content KVScontent) error
}
