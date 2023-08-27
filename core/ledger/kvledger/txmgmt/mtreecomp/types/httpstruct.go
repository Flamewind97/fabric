package types

type MerkleRootResponse struct {
	Data           []byte `json:"data"`
	LastCommitHash []byte `json:"lastCommitHash"`
}

type SignedMerkleRootResponse struct {
	SerializedMerkleRootResponse []byte `json:"serializedMerkleRootResponse"`
	Signature                    []byte `json:"signature"`
}
