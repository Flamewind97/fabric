package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/types"
)

func main() {
	response, err := http.Get("http://localhost:8080/merkleRoot?namespace=test")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code %d", response.StatusCode)
	}

	var signedResponse types.SignedMerkleRootResponse
	err = json.Unmarshal(body, &signedResponse)
	if err != nil {
		log.Fatal(err)
	}

	var merkleRootResponse types.MerkleRootResponse
	err = json.Unmarshal(signedResponse.SerializedMerkleRootResponse, &merkleRootResponse)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Merkle Data:", string(merkleRootResponse.Data))
	fmt.Println("Last Commit Hash:", string(merkleRootResponse.LastCommitHash))
	fmt.Println("Signature:", string(signedResponse.Signature))
}
