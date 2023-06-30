package mtreecomp

import (
	"sync"

	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/mtreeimpl"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
)

type MerkleTreeComponent struct {
	mapMTree      map[string]types.MerkleTree
	newMerkleTree func(contents []types.KVScontent) (types.MerkleTree, error)
}

var once sync.Once
var merkleComponentInstance *MerkleTreeComponent

func GetMerkleTreeComponent() (*MerkleTreeComponent, error) {
	var err error
	err = nil
	if merkleComponentInstance == nil {
		once.Do(
			func() {
				merkleComponentInstance, err = NewMerkleTreeComponent()
				ServeMerkle("0.0.0.0:12345", merkleComponentInstance)
			})
	}
	return merkleComponentInstance, err
}

func NewMerkleTreeComponent() (*MerkleTreeComponent, error) {
	return &MerkleTreeComponent{
		mapMTree:      make(map[string]types.MerkleTree),
		newMerkleTree: mtreeimpl.NewMerkleTreeCbergoon,
	}, nil
}

func (m *MerkleTreeComponent) GetMerkleRoot(ns string) ([]byte, error) {
	mtree, found := m.mapMTree[ns]
	if found != true {
		return nil, nil
	}
	return mtree.GetMerkleRoot(), nil
}

func (m *MerkleTreeComponent) GetMerklePath(ns string, content types.KVScontent) ([]types.MerklePath, error) {
	mtree, found := m.mapMTree[ns]
	if found != true {
		return nil, nil
	}
	return mtree.GetMerklePath(content)
}

func (m *MerkleTreeComponent) VerifyContent(ns string, content types.KVScontent) (bool, error) {
	mtree, found := m.mapMTree[ns]
	if found != true {
		return found, nil
	}
	return mtree.VerifyContent(content)
}

func (m *MerkleTreeComponent) ApplyUpdates(batch *statedb.UpdateBatch, height *version.Height) error {
	namespaces := batch.GetUpdatedNamespaces()
	for _, ns := range namespaces {
		mtree, found := m.mapMTree[ns]

		if !found {
			newMTree, err := m.newMerkleTree([]types.KVScontent{})
			if err != nil {
				return err
			}
			mtree = newMTree
			m.mapMTree[ns] = mtree
		}

		updates := batch.GetUpdates(ns)
		for k, vv := range updates {
			c := types.KVScontent{
				Key:   k,
				Value: vv.Value,
			}
			if vv.Value == nil {
				mtree.Delete(c)
			} else {
				mtree.Add(c)
			}
		}
	}
	// TODO: record a savepoint at a given height.
	return nil
}
