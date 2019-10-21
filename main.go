package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "fyfe.io/merkle"
	"google.golang.org/grpc"
)

func main() {
	var port, chunkSize int
	var filename string
	flag.IntVar(&port, "port", 9999, "Listen port")
	flag.IntVar(&chunkSize, "chunksize", 32, "Chunk size in bytes")
	flag.StringVar(&filename, "filename", "", "file to transfer")

	flag.Parse()

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	tree := NewMerkleTree(chunkSize, filename)
	tree.Read(file)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	server := &merkleServer{tree}
	pb.RegisterMerkleServer(grpcServer, server)

	fmt.Printf("%X %X", tree.Chunks[0], tree.Chunks[7])

	grpcServer.Serve(lis)
}
