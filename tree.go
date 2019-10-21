package main

import (
	"encoding/json"
	"fmt"
	"io"

	pb "fyfe.io/merkle"
)

// Node is a tree node
type Node struct {
	parent      *Node
	left, right *Node
	hash        string
}

// ProofNode is a node of merkle proof
type ProofNode struct {
	left bool
	hash string
}

// MerkleProof is the proof required to prove a node
type MerkleProof []ProofNode

// MarshalJSON is the custom JSON implementation
func (n Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Left  *Node `json:"left"`
		Right *Node `json:"right"`
		Hash  string
	}{
		n.left,
		n.right,
		n.hash,
	})
}

// calculate calculates the hash from the children
func (n *Node) calculate() {
	n.hash = pb.Hash([]byte(n.left.hash + n.right.hash))
	n.left.parent = n
	n.right.parent = n
}

// Sibling gets the node's sibling
func (n *Node) Sibling() *Node {
	if n.parent == nil {
		return nil
	}
	if n.parent.left.hash == n.hash {
		return n.parent.right
	}
	return n.parent.left
}

// Uncle gets the node's parent's sibling
func (n *Node) Uncle() *Node {
	if n.parent == nil {
		return nil
	}
	return n.parent.Sibling()
}

// MerkleTree is a tree specific for file sharing
type MerkleTree struct {
	chunkSize int
	Filename  string
	Chunks    [][]byte
	Nodes     [][]*Node
	rootNode  *Node
	height    int
}

// NewMerkleTree creates a new MerkleTree
func NewMerkleTree(chunkSize int, filename string) *MerkleTree {
	return &MerkleTree{
		chunkSize,
		filename,
		[][]byte{},
		[][]*Node{},
		nil,
		0,
	}
}

// Read reads data in to the tree
func (t *MerkleTree) Read(reader io.Reader) {
	buf := make([]byte, t.chunkSize)

	fmt.Println("Chunking file:")
	t.Nodes = append(t.Nodes, []*Node{})
	i := 0
	for {
		read, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		hash := pb.Hash(buf[:read])

		fmt.Printf("#%03d %s\n", i, hash)
		newData := make([]byte, read)
		copy(newData, buf[:read])
		fmt.Printf("chunk %d: %x\n", i, newData)
		t.Chunks = append(t.Chunks, newData)
		t.Nodes[0] = append(t.Nodes[0], &Node{hash: hash})
		i++
	}
	fmt.Println("Done!")
	fmt.Printf("Building merkle tree: height ")
	height := 0
	for t.rootNode == nil {
		fmt.Printf("%d ", height)
		if len(t.Nodes[height]) < 2 {
			t.rootNode = t.Nodes[height][0]
		}
		nextHeight := make([]*Node, 0)
		for i := 0; i < int(len(t.Nodes[height])/2)*2; i += 2 {
			newNode := &Node{left: t.Nodes[height][i], right: t.Nodes[height][i+1]}
			newNode.calculate()
			nextHeight = append(nextHeight, newNode)
		}
		if len(t.Nodes[height])%2 != 0 {
			nextHeight = append(nextHeight, t.Nodes[height][len(t.Nodes[height])-1])
		}
		t.Nodes = append(t.Nodes, nextHeight)
		height++
	}
	fmt.Printf("\n")
	t.height = height
	fmt.Printf("Root %s\n\n", t.rootNode.hash)
}

// GetProof gets a list of proofs
func (t *MerkleTree) GetProof(idx int) *pb.Proof {
	proof := pb.Proof{Nodes: []*pb.Proof_ProofNode{}, MerkleRoot: t.rootNode.hash}

	nextProof := t.Nodes[0][idx].Sibling()
	for {
		side := pb.Proof_ProofNode_LEFT
		if nextProof.parent.right.hash == nextProof.hash {
			side = pb.Proof_ProofNode_RIGHT
		}
		proof.Nodes = append(proof.Nodes, &pb.Proof_ProofNode{Side: side, Hash: nextProof.hash})
		if nextProof.Uncle() == nil {
			// 	proof.Nodes = append(proof.Nodes, &pb.Proof_ProofNode{Side: side, Hash: nextProof.parent.hash})
			break
		}
		nextProof = nextProof.Uncle()
	}
	return &proof
}

// GetProof returns a list of hashes and whether to place your computed hash on the left
// The first item in the hash is for the sibling at height 0, then for the sibling of the computed hash
// The last item in the hash is the root hash and should be compared against the computed root hash.
func GetProof(n *Node) MerkleProof {
	proof := MerkleProof{}
	nextProof := n.Sibling()
	for {
		left := nextProof.parent.right.hash == nextProof.hash
		proof = append(proof, ProofNode{left, nextProof.hash})
		if nextProof.Uncle() == nil {
			proof = append(proof, ProofNode{left, nextProof.parent.hash})
			break
		}
		nextProof = nextProof.Uncle()
	}
	return proof
}

// Prove proves that the a hash is correct for the given proof
func (p MerkleProof) Prove(ha string) bool {
	rootHash := p[len(p)-1].hash
	for i := 0; i < len(p)-1; i++ {
		if p[i].left {
			ha = pb.Hash([]byte(ha + p[i].hash))
		} else {
			ha = pb.Hash([]byte(p[i].hash + ha))
		}
		fmt.Printf("#%03d: %s\n", i, ha)
	}

	return ha == rootHash
}
