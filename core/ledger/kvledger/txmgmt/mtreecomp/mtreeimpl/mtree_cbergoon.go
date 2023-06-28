package mtreeimpl

import (
	"bytes"
	"errors"

	mtc "github.com/cbergoon/merkletree"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

type mtcContent struct {
	kvs *types.KVScontent
}

// CalculateHash hashes the values of a TestContent
func (m *mtcContent) CalculateHash() ([]byte, error) {
	return m.kvs.CalculateHash()
}

// Equals tests for equality of two Contents
func (m *mtcContent) Equals(other mtc.Content) (bool, error) {
	otherTC, ok := other.(*mtcContent)
	if !ok {
		return false, errors.New("value is not of type mtcContent")
	}
	h1, err := m.CalculateHash()
	if err != nil {
		return false, err
	}
	h2, err := otherTC.CalculateHash()
	if err != nil {
		return false, err
	}

	return bytes.Equal(h1, h2), nil
}

type MerkleTreeCbergoon struct {
	mtc        *mtc.MerkleTree
	ContentMap map[string]mtc.Content
}

func NewMerkleTreeCbergoon(contents []types.KVScontent) (types.MerkleTree, error) {
	mtree := &MerkleTreeCbergoon{
		ContentMap: map[string]mtc.Content{},
	}

	// Because MT cannot handle empty contents
	if len(contents) == 0 {
		mtree.ContentMap["INIT"] = &mtcContent{
			kvs: &types.KVScontent{
				Key:   "INIT",
				Value: []byte("INIT"),
			},
		}
	}

	for _, c := range contents {
		mtcc := mtree.convertContent(c)
		mtree.ContentMap[c.Key] = mtcc
	}
	tree, err := mtc.NewTree(mtree.buildContentFromMap())
	if err != nil {
		return nil, err
	}
	mtree.mtc = tree
	return mtree, nil
}

func (m *MerkleTreeCbergoon) GetMerkleRoot() []byte {
	return m.mtc.MerkleRoot()
}

func (m *MerkleTreeCbergoon) GetMerklePath(content types.KVScontent) ([]types.MerklePath, error) {
	mtcc, found := m.ContentMap[content.Key]
	if found == false {
		return []types.MerklePath{}, nil
	}

	mtcPath, index, err := m.mtc.GetMerklePath(mtcc)

	mpath := make([]types.MerklePath, 0)
	for i, p := range mtcPath {
		mp := types.MerklePath{
			Path: p,
			Pos:  bool(index[i] != 0),
		}
		mpath = append(mpath, mp)
	}
	return mpath, err
}

func (m *MerkleTreeCbergoon) VerifyContent(content types.KVScontent) (bool, error) {
	mtcc, found := m.ContentMap[content.Key]
	if found == false {
		return false, nil
	}

	valid, err := mtcc.Equals(m.convertContent(content))
	if err != nil || valid != true {
		return valid, err
	}

	return m.mtc.VerifyContent(mtcc)
}

func (m *MerkleTreeCbergoon) Add(content types.KVScontent) error {
	_, found := m.ContentMap[content.Key]
	if found {
		return m.Update(content)
	}

	mtcc := m.convertContent(content)
	m.ContentMap[content.Key] = mtcc

	t, err := mtc.NewTree(m.buildContentFromMap())
	m.mtc = t
	return err
}

func (m *MerkleTreeCbergoon) Delete(content types.KVScontent) error {
	_, found := m.ContentMap[content.Key]
	if found == true {
		delete(m.ContentMap, content.Key)
		t, err := mtc.NewTree(m.buildContentFromMap())
		m.mtc = t
		return err
	}
	return nil
}

func (m *MerkleTreeCbergoon) Update(content types.KVScontent) error {
	m.ContentMap[content.Key] = m.convertContent(content)
	t, err := mtc.NewTree(m.buildContentFromMap())
	m.mtc = t
	return err
}

func (m *MerkleTreeCbergoon) convertContent(c types.KVScontent) mtc.Content {
	return &mtcContent{kvs: &c}
}

func (m *MerkleTreeCbergoon) buildContentFromMap() []mtc.Content {
	mtcContents := make([]mtc.Content, 0)
	for _, c := range m.ContentMap {
		mtcContents = append(mtcContents, c)
	}

	return mtcContents
}
