package mtreeimpl

import (
	"crypto/sha256"
	"hash"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

// Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
// relationships, a hash, the types.KVScontent stored if it is a leaf, and other metadata.
type Node struct {
	Tree   *MerkleTree
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte
	C      types.KVScontent
}

// calculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node) calculateNodeHash() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
// a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root         *Node
	merkleRoot   []byte
	Leafs        map[string]*Node
	hashStrategy func() hash.Hash
}

// NewTree creates a new Merkle Tree using the types.KVScontent cs using the provided hash
// strategy. Note that the hash type used in the type that implements the types.KVScontent interface must
// match the hash type profided to the tree.
// If hashStrategy is nil then use default hash sha256
func NewTree(cs []types.KVScontent, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	if hashStrategy == nil {
		hashStrategy = sha256.New
	}
	t := &MerkleTree{
		hashStrategy: hashStrategy,
	}
	root, err := buildTree(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	// t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

// Implement GetMerkleRoot
func (m *MerkleTree) GetMerkleRoot() []byte {
	return m.merkleRoot
}

// Implement GetMerklePath
func (m *MerkleTree) GetMerklePath(content types.KVScontent) ([]types.MerklePath, error) {
	panic("not implemented")
}

// Implement VerifyContent
func (m *MerkleTree) VerifyContent(content types.KVScontent) (bool, error) {
	panic("not implemented")
}

// Implement Add
func (m *MerkleTree) Add(content types.KVScontent) error {
	panic("not implemented")
}

// Implement Delete
func (m *MerkleTree) Delete(content types.KVScontent) error {
	panic("not implemented")
}

// Update
func (m *MerkleTree) Update(content types.KVScontent) error {
	panic("not implemented")
}

// buildTree is a helper function that for a given set of Contents, generates a
// corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
// Returns an error if cs contains no Contents.
func buildTree(cs []types.KVScontent, t *MerkleTree) (*Node, error) {
	panic("not implemented")
}

// buildIntermediate is a helper function that for a given list of leaf nodes, constructs
// the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node, t *MerkleTree) (*Node, error) {
	panic("not implemented")
}
