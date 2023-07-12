package mtreeimpl

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

// Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
// relationships, a hash, the types.KVScontent stored if it is a leaf, and other metadata.
type Node struct {
	Tree      *MerkleTree
	Parent    *Node
	Left      *Node
	Right     *Node
	leaf      bool
	Hash      []byte
	PrefixKey []byte
	PrefixLen int
	C         *types.KVScontent
}

// calculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node) calculateNodeHash() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}

	h := n.Tree.hashStrategy()

	leftBytes := []byte{}
	rightBytes := []byte{}

	if n.Left != nil {
		leftBytes = n.Left.Hash
	}
	if n.Right != nil {
		rightBytes = n.Right.Hash
	}

	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
// a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root         *Node
	Leafs        map[string]*Node
	hashStrategy func() hash.Hash
}

// NewTree creates a new Merkle Tree using the types.KVScontent cs using the provided hash
// strategy. Note that the hash type used in the type that implements the types.KVScontent interface must
// match the hash type profided to the tree.
// If hashStrategy is nil then use default hash sha256
func NewTree(cs []types.KVScontent, hashStrategy func() hash.Hash) (types.MerkleTree, error) {
	if hashStrategy == nil {
		hashStrategy = sha256.New
	}
	t := &MerkleTree{
		hashStrategy: hashStrategy,
		Leafs:        map[string]*Node{},
	}
	root := &Node{
		Tree:      t,
		Parent:    nil,
		Left:      nil,
		Right:     nil,
		leaf:      false,
		Hash:      hashStrategy().Sum([]byte("init")),
		PrefixKey: hashStrategy().Sum([]byte("init")),
		PrefixLen: 0,
	}
	t.Root = root

	err := buildTree(cs, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Implement GetMerkleRoot
func (m *MerkleTree) GetMerkleRoot() []byte {
	return m.Root.Hash
}

// Implement GetMerklePath
func (m *MerkleTree) GetMerklePath(content types.KVScontent) ([]types.MerklePath, error) {
	curNode, found := m.Leafs[content.Key]
	if !found {
		return []types.MerklePath{}, nil
	}

	mpath := make([]types.MerklePath, 0)
	for curNode.Parent != nil {
		parentNode := curNode.Parent
		mp := types.MerklePath{
			Path: []byte{},
		}
		if curNode == parentNode.Left {
			if parentNode.Right != nil {
				mp.Path = parentNode.Right.Hash
			}
			mp.Pos = true
		} else {
			if parentNode.Left != nil {
				mp.Path = parentNode.Left.Hash
			}
			mp.Pos = false
		}
		mpath = append(mpath, mp)
		curNode = parentNode
	}
	return mpath, nil
}

// Implement VerifyContent
func (m *MerkleTree) VerifyContent(content types.KVScontent) (bool, error) {
	curNode, found := m.Leafs[content.Key]
	if !found {
		return false, nil
	}

	// check if hashValue equal curNode hashValue
	hashValue, err := content.CalculateHash()
	if err != nil {
		return false, nil
	}
	if !bytes.Equal(curNode.Hash, hashValue) {
		return false, nil
	}

	// verify path
	for curNode.Parent != nil {
		h := m.hashStrategy()
		curNode = curNode.Parent
		leftBytes := []byte{}
		rightBytes := []byte{}

		if curNode.Left != nil {
			leftBytes, err = curNode.Left.calculateNodeHash()
			if err != nil {
				return false, err
			}
		}
		if curNode.Right != nil {
			rightBytes, err = curNode.Right.calculateNodeHash()
			if err != nil {
				return false, err
			}
		}

		if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
			return false, err
		}
		if !bytes.Equal(h.Sum(nil), curNode.Hash) {
			return false, nil
		}
	}
	return true, nil
}

// Implement Add
func (m *MerkleTree) Add(content types.KVScontent) error {
	_, found := m.Leafs[content.Key]
	if found {
		err := m.Update(content)
		if err != nil {
			return err
		}
	}

	// create new node
	hashKey := m.hashStrategy().Sum([]byte(content.Key))
	hashValue, err := content.CalculateHash()
	if err != nil {
		return err
	}
	newContent := content
	newNode := &Node{
		Tree:      m,
		Parent:    nil,
		Left:      nil,
		Right:     nil,
		leaf:      true,
		Hash:      hashValue,
		PrefixKey: hashKey,
		PrefixLen: 0,
		C:         &newContent,
	}
	m.Leafs[content.Key] = newNode

	// traverse from m root
	// 0~127 left node, 128~255 right node
	if newNode.PrefixKey[0] < 128 {
		if m.Root.Left == nil {
			m.Root.Left = newNode
			newNode.Parent = m.Root
		} else {
			m.addNode(m.Root.Left, newNode)
		}
	} else {
		if m.Root.Right == nil {
			m.Root.Right = newNode
			newNode.Parent = m.Root
		} else {
			m.addNode(m.Root.Right, newNode)
		}
	}
	m.Root.Hash, err = m.Root.calculateNodeHash()
	return err
}

func (m *MerkleTree) addNode(curNode, newNode *Node) error {
	var err error

	prefixLen := BitsSimilarPrefixLength(curNode.PrefixKey, newNode.PrefixKey)
	// if is leaf or reach the point where unequal prefix, create a parent node in the middle.
	if curNode.leaf || prefixLen < curNode.PrefixLen {
		// build a parent node.
		oriParentNode := curNode.Parent
		newParentNode := &Node{
			Tree:      m,
			Parent:    oriParentNode,
			Left:      nil,
			Right:     nil,
			leaf:      false,
			Hash:      nil,
			PrefixKey: curNode.PrefixKey,
			PrefixLen: prefixLen,
			C:         nil,
		}

		// update original parent
		if oriParentNode.Left == curNode {
			oriParentNode.Left = newParentNode
		} else {
			oriParentNode.Right = newParentNode
		}

		// if smaller, curNode on left, newNode on right, otherwise.
		if bytes.Compare(curNode.PrefixKey, newNode.PrefixKey) < 0 {
			newParentNode.Left = curNode
			newParentNode.Right = newNode
		} else {
			newParentNode.Left = newNode
			newParentNode.Right = curNode
		}
		curNode.Parent = newParentNode
		newNode.Parent = newParentNode

		newParentNode.Hash, err = newParentNode.calculateNodeHash()
		return err
	}
	// keep traverse the node
	prefixLeft := BitsSimilarPrefixLength(curNode.Left.PrefixKey, newNode.PrefixKey)
	prefixRight := BitsSimilarPrefixLength(curNode.Right.PrefixKey, newNode.PrefixKey)
	// there will no be equal case here, by theory unless there is a bug.
	if prefixLeft > prefixRight {
		err = m.addNode(curNode.Left, newNode)
	} else {
		err = m.addNode(curNode.Right, newNode)
	}
	if err != nil {
		return err
	}
	curNode.Hash, err = curNode.calculateNodeHash()
	return err
}

// Implement Delete
func (m *MerkleTree) Delete(content types.KVScontent) error {
	delNode, found := m.Leafs[content.Key]
	if found {
		delete(m.Leafs, content.Key)

		parentNode := delNode.Parent

		// dereference delnode
		delNode.Parent = nil

		// root case
		if parentNode.Parent == nil {
			if delNode == parentNode.Left {
				parentNode.Left = nil
			} else {
				parentNode.Right = nil
			}
			parentNode.Hash, _ = parentNode.calculateNodeHash()
		} else { // non root case
			replaceNode := parentNode.Left
			if delNode == parentNode.Left {
				replaceNode = parentNode.Right
			}
			grandParentNode := parentNode.Parent
			replaceNode.Parent = grandParentNode

			if grandParentNode.Left == parentNode {
				grandParentNode.Left = replaceNode
			} else {
				grandParentNode.Right = replaceNode
			}
			// dereference parentNode
			parentNode.Parent = nil
			parentNode.Left = nil
			parentNode.Right = nil

			// update parent hash
			for replaceNode.Parent != nil {
				replaceNode = replaceNode.Parent
				replaceNode.Hash, _ = parentNode.calculateNodeHash()
			}
		}
	}
	return nil
}

// Implement Update
func (m *MerkleTree) Update(content types.KVScontent) error {
	var err error
	updateNode, found := m.Leafs[content.Key]
	if !found {
		return m.Add(content)
	}
	updateContent := content
	updateNode.C = &updateContent
	updateNode.Hash, err = updateNode.calculateNodeHash()
	if err != nil {
		return err
	}

	for updateNode.Parent != nil {
		updateNode = updateNode.Parent
		updateNode.Hash, err = updateNode.calculateNodeHash()
		if err != nil {
			return err
		}
	}
	return nil
}

// buildTree is a helper function that for a given set of Contents, generates a
// corresponding tree and returns the possible error.
func buildTree(cs []types.KVScontent, t *MerkleTree) error {
	for _, c := range cs {
		err := t.Add(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func BitsSimilarPrefixLength(h1, h2 []byte) int {
	n := 0
	for i := range h1 {
		xor := h1[i] ^ h2[i]
		if xor != 0 {
			n = i*8 + leadingZeros(xor)
			break
		}
	}
	return n
}

func leadingZeros(b byte) int {
	count := 0
	for i := 7; i >= 0; i-- {
		if (b & (1 << uint(i))) != 0 {
			break
		}
		count++
	}
	return count
}

func printMerkleTree(tree *MerkleTree) {
	printNode(tree.Root, "", true)
}

func printNode(node *Node, prefix string, isTail bool) {
	fmt.Printf("%s%s%s\n", prefix, getPrefix(isTail), getNodeString(node))

	if node.Left != nil {
		printNode(node.Left, prefix+getPrefix(isTail)+"   ", false)
	}

	if node.Right != nil {
		printNode(node.Right, prefix+getPrefix(isTail)+"   ", true)
	}
}

func getPrefix(isTail bool) string {
	if isTail {
		return "└── "
	}
	return "├── "
}

func getNodeString(node *Node) string {
	if node.Parent == nil {
		return fmt.Sprintf("Root: %x", node.Hash)
	} else if node.leaf {
		return fmt.Sprintf("Leaf: key: %s, value: %x", node.C.Key, node.C.Value)
	}
	return fmt.Sprintf("Node: %x", node.Hash)
}
