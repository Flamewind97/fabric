package mtreeimpl

import (
	"bytes"
	"crypto/sha256"
	"log"
	"testing"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
	"github.com/stretchr/testify/require"
)

func Setup(t *testing.T) (types.MerkleTree, []types.KVScontent) {
	// Build list of Content to build tree
	var list []types.KVScontent
	list = append(list, types.KVScontent{Key: "Hello", Value: []byte("hello")})
	list = append(list, types.KVScontent{Key: "Hi", Value: []byte("hi")})
	list = append(list, types.KVScontent{Key: "Hey", Value: []byte("hey")})
	list = append(list, types.KVScontent{Key: "Hola", Value: []byte("hey")})

	// Create a new Merkle Tree from the list of Content
	// tree, err := NewMerkleTreeCbergoon(list)
	tree, err := NewTree(list, nil)
	require.NoError(t, err)

	PrintMerkleTree(tree.(*MerkleTree))
	return tree, list
}

func TestAdd(t *testing.T) {
	tree, _ := Setup(t)

	// Verify a non exist content
	newContent := types.KVScontent{Key: "Nihao", Value: []byte("hello")}
	vc, err := tree.VerifyContent(newContent)
	require.NoError(t, err)
	require.False(t, vc)

	// check for merkle paths
	tree.Add(newContent)
	mr := tree.GetMerkleRoot()
	merklepath, err := tree.GetMerklePath(newContent)
	require.NoError(t, err)
	log.Printf("update %v, mr: %x, mp: %v", newContent, mr, merklepath)
	res, err := VerifyMerklePath(newContent, merklepath, mr, sha256.New)
	require.NoError(t, err)
	require.True(t, res)

	PrintMerkleTree(tree.(*MerkleTree))

	// add same key twice
	newContent2 := types.KVScontent{Key: "Nihao", Value: []byte("hallo2")}
	tree.Add(newContent2)
	mr = tree.GetMerkleRoot()
	merklepath, err = tree.GetMerklePath(newContent2)
	require.NoError(t, err)
	log.Printf("update %v, mr: %x, mp: %v", newContent2, mr, merklepath)
	res, err = VerifyMerklePath(newContent2, merklepath, mr, sha256.New)
	require.NoError(t, err)
	require.True(t, res)
	PrintMerkleTree(tree.(*MerkleTree))
}

func TestDel(t *testing.T) {
	tree, list := Setup(t)

	mr := tree.GetMerkleRoot()

	// deleting an non exist contents
	deleteContent := types.KVScontent{Key: "Nihao", Value: []byte("hi")}
	err := tree.Delete(deleteContent)
	require.NoError(t, err)
	mr2 := tree.GetMerkleRoot()
	require.Equal(t, mr, mr2)

	// deleting an exist contents
	err = tree.Delete(list[2])
	require.NoError(t, err)
	mr2 = tree.GetMerkleRoot()
	require.False(t, bytes.Equal(mr, mr2))

	// check for merkle path
	merklepath, err := tree.GetMerklePath(list[2])
	require.NoError(t, err)
	res, err := VerifyMerklePath(list[2], merklepath, mr2, sha256.New)
	require.NoError(t, err)
	require.False(t, res)
	vc, err := tree.VerifyContent(list[2])
	require.NoError(t, err)
	require.False(t, vc)
}

func TestUpdate(t *testing.T) {
	tree, _ := Setup(t)

	mr := tree.GetMerkleRoot()

	// update an exist contents
	updateContent := types.KVScontent{Key: "Hola", Value: []byte("hi")}
	err := tree.Update(updateContent)
	require.NoError(t, err)
	mr2 := tree.GetMerkleRoot()
	require.False(t, bytes.Equal(mr, mr2))

	// verify content
	vc, err := tree.VerifyContent(updateContent)
	require.NoError(t, err)
	require.True(t, vc)

	// verify merkle path
	merklepath, err := tree.GetMerklePath(updateContent)
	require.NoError(t, err)
	res, err := VerifyMerklePath(updateContent, merklepath, mr2, sha256.New)
	require.NoError(t, err)
	require.True(t, res)
}

func TestMultiple(t *testing.T) {
	tree, list := Setup(t)

	//Get the Merkle Root of the tree
	mr := tree.GetMerkleRoot()
	log.Printf("M root: %x\n", mr)

	//Verify a specific content in in the tree
	for _, x := range list {
		vc, err := tree.VerifyContent(x)
		require.NoError(t, err)
		log.Println("Verify Content:", x, ", result:", vc)
		require.True(t, vc)
	}
	//Verify a non exist content
	newContent := types.KVScontent{Key: "Nihao", Value: []byte("nihao")}

	vc, err := tree.VerifyContent(newContent)
	require.NoError(t, err)
	require.False(t, vc)

	tree.Add(newContent)
	mr = tree.GetMerkleRoot()
	log.Printf("New M root: %x\n", mr)

	vc, err = tree.VerifyContent(newContent)
	require.NoError(t, err)
	require.True(t, vc)

	// get merkle path
	merklepath, err := tree.GetMerklePath(newContent)
	log.Println("GetMerklePath for ", newContent)
	for _, p := range merklepath {
		log.Printf("(%v, %x)\n", p.Pos, p.Path)
	}
	require.NoError(t, err)

	// verify merkle path
	res, err := VerifyMerklePath(newContent, merklepath, mr, sha256.New)
	require.NoError(t, err)
	require.True(t, res)

	// test update
	updateContent := types.KVScontent{Key: "Hello", Value: []byte("hallo")}
	err = tree.Update(updateContent)
	require.NoError(t, err)
	mr = tree.GetMerkleRoot()
	merklepath, err = tree.GetMerklePath(updateContent)
	require.NoError(t, err)
	log.Printf("update %v, mr: %x, mp: %v", updateContent, mr, merklepath)
	res, err = VerifyMerklePath(updateContent, merklepath, mr, sha256.New)
	require.NoError(t, err)
	require.True(t, res)

	// test delete
	deleteContent := types.KVScontent{Key: "Hi", Value: []byte("hi")}
	err = tree.Delete(deleteContent)
	require.NoError(t, err)
	mr = tree.GetMerkleRoot()
	merklepath, err = tree.GetMerklePath(deleteContent)
	require.NoError(t, err)
	log.Printf("delete %v, mr: %x, mp: %v", deleteContent, mr, merklepath)
	res, err = VerifyMerklePath(deleteContent, merklepath, mr, sha256.New)
	require.NoError(t, err)
	require.False(t, res)
	vc, err = tree.VerifyContent(deleteContent)
	require.NoError(t, err)
	require.False(t, vc)
}
