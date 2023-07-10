package mtreecomp

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

type MerkleRest struct {
	mtci *MerkleTreeComponent
}

func (s *MerkleRest) signResponse([]byte) ([]byte, error) {
	// TODO, signed the merkleRoot instead of random bytes.
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func (s *MerkleRest) GetMerkleRoot(namespace string) (*types.SignedMerkleRootResponse, error) {
	mroot, err := s.mtci.GetMerkleRoot(namespace)
	fmt.Printf("MerkleServer get merkle root namespace: %s, mroot: %x\n", namespace, mroot)
	if err != nil {
		return nil, err
	}
	// TODO, change lastCommitHash to last txn hash instead of mroot.
	merkleRootResponse := &types.MerkleRootResponse{Data: mroot, LastCommitHash: mroot}
	bytesResponse, err := json.Marshal(merkleRootResponse)
	if err != nil {
		return nil, err
	}

	signature, err := s.signResponse(bytesResponse)
	if err != nil {
		return nil, err
	}
	signedResponse := &types.SignedMerkleRootResponse{
		SerializedMerkleRootResponse: bytesResponse,
		Signature:                    signature,
	}

	return signedResponse, nil
}

func (s *MerkleRest) getMerkleRootHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	namespace := r.URL.Query().Get("namespace")

	signedResponse, err := s.GetMerkleRoot(namespace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Convert the response to JSON
	jsonResponse, err := json.Marshal(signedResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write the JSON response
	w.Write(jsonResponse)
}

func ServeMerkle(listenAddress string, mtci *MerkleTreeComponent) {
	s := &MerkleRest{mtci: mtci}
	http.HandleFunc("/merkleRoot", s.getMerkleRootHandler)
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
