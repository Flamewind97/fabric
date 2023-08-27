package mtreecomp

import (
	"fmt"
	"hash"
	"strconv"
	"strings"
	"sync"

	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/mtreeimpl"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/spf13/viper"
)

type MerkleTreeComponent struct {
	mapMTree      map[string]types.MerkleTree
	newMerkleTree func(contents []types.KVScontent, hashStrategy func() hash.Hash) (types.MerkleTree, error)
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

				// serve http server.
				peerAddr := viper.GetString("peer.listenAddress")
				peerPort := strings.Split(peerAddr, ":")[1]
				portnum, _ := strconv.Atoi(peerPort)
				portnum += 1000
				httpAddr := "0.0.0.0:" + strconv.Itoa(portnum)
				go ServeMerkle(httpAddr, merkleComponentInstance)
			})
	}
	return merkleComponentInstance, err
}

func NewMerkleTreeComponent() (*MerkleTreeComponent, error) {
	return &MerkleTreeComponent{
		mapMTree:      make(map[string]types.MerkleTree),
		newMerkleTree: mtreeimpl.NewTree,
		// newMerkleTree: mtreeimpl.NewMerkleTreeCbergoon,
	}, nil
}

func (m *MerkleTreeComponent) GetMerkleRoot(ns string) ([]byte, error) {
	mtree, found := m.mapMTree[ns]
	if !found {
		return nil, nil
	}
	return mtree.GetMerkleRoot(), nil
}

func (m *MerkleTreeComponent) GetMerklePath(ns string, content types.KVScontent) ([]types.MerklePath, error) {
	mtree, found := m.mapMTree[ns]
	if !found {
		return nil, nil
	}
	return mtree.GetMerklePath(content)
}

func (m *MerkleTreeComponent) VerifyContent(ns string, content types.KVScontent) (bool, error) {
	mtree, found := m.mapMTree[ns]
	if !found {
		return found, nil
	}
	return mtree.VerifyContent(content)
}

func (m *MerkleTreeComponent) ApplyUpdates(batch *statedb.UpdateBatch, height *version.Height) error {
	fmt.Printf("=== MerkleTreeComponent, ApplyUpdates, height=%v, batch=%v ===\n", height, batch)
	namespaces := batch.GetUpdatedNamespaces()
	for _, ns := range namespaces {
		mtree, found := m.mapMTree[ns]

		if !found {
			newMTree, err := m.newMerkleTree([]types.KVScontent{}, nil)
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
