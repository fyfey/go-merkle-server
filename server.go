package main

import (
	"context"
	"errors"

	pb "fyfe.io/merkle"
)

type merkleServer struct {
	Tree *MerkleTree
}

// GetMetedata gets the file's metadata
func (s *merkleServer) GetMetadata(context.Context, *pb.Empty) (*pb.Metadata, error) {
	return &pb.Metadata{
		Filename:  s.Tree.Filename,
		Parts:     int32(len(s.Tree.Chunks)),
		ChunkSize: int32(s.Tree.chunkSize),
	}, nil
}

// GetPart returns a given part
func (s *merkleServer) GetPart(ctx context.Context, in *pb.PartRequest) (*pb.Part, error) {
	data := s.Tree.Chunks[in.Idx]
	if len(data) == 0 {
		return nil, errors.New("Part does not exist")
	}
	return &pb.Part{
		Idx:   in.Idx,
		Data:  data,
		Proof: s.Tree.GetProof(int(in.Idx)),
	}, nil
}
