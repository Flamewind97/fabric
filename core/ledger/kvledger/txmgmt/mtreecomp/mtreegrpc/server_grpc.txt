package mtreecomp

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/mtreecomp/mtreegrpc"
	"google.golang.org/grpc"
)

type MerkleServer struct {
	pb.UnimplementedMerkleServiceServer
	mtci *MerkleTreeComponent
}

func (s *MerkleServer) signResponse([]byte) ([]byte, error) {
	// TODO, signed the merkleRoot instead of random bytes.
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func (s *MerkleServer) GetMerkleRoot(ctx context.Context, req *pb.MerkleRootRequest) (*pb.SignedMerkleRootResponse, error) {
	namespace := req.Namespace
	mroot, err := s.mtci.GetMerkleRoot(namespace)
	fmt.Printf("MerkleServer get merkle root namespace: %s, mroot: %x\n", namespace, mroot)
	if err != nil {
		return nil, err
	}
	// TODO, change lastCommitHash to last txn hash instead of mroot.
	merkleRootResponse := &pb.MerkleRootResponse{Data: mroot, LastCommitHash: mroot}
	bytesResponse, err := proto.Marshal(merkleRootResponse)
	if err != nil {
		return nil, err
	}

	signature, err := s.signResponse(bytesResponse)
	signedResponse := &pb.SignedMerkleRootResponse{
		SerializedMerkleRootResponse: bytesResponse,
		Signature:                    signature,
	}

	return signedResponse, nil
}

func ServeMerkle(port string, mtci *MerkleTreeComponent) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterMerkleServiceServer(s, &MerkleServer{mtci: mtci})
	fmt.Println("Server listening on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
