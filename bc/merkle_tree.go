package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	RootNode *Node
	Leafs    []*Node
}

// Node represents a Merkle tree node
type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	Hash   []byte
}

const (
	leftNode = iota
	rightNode
)

// MerkleProof represents way to prove element inclusion on the merkle tree
type MerkleProof struct {
	proof [][]byte
	index []int64
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	mrkltree := &MerkleTree{}
	if len(data) == 0 {
		panic("No merkle tree nodes")
	}
	if len(data) == 1 {
		newnode := NewMerkleNode(nil, nil, data[0])
		mrkltree.RootNode = newnode
		return mrkltree
	}
	// run the first level algo.
	mrkltree.firstLevelNodes(data)
	// run the rest levels
	i := 1
	treeSize := len(mrkltree.Leafs)
	for mrkltree.RootNode == nil {
		if mrkltree.Leafs[i-1].Parent == nil && mrkltree.Leafs[i].Parent == nil {
			// insert new nodes at the end of the slice
			mrkltree.insertNode(i)
		}
		if i == (treeSize - 1) {
			treeSize = len(mrkltree.Leafs)
			if treeSize == i+2 { // size of the "old list" is "old list + 1"
				root := mrkltree.Leafs[treeSize-1] // we have one node left, which is root
				mrkltree.RootNode = root
				return mrkltree
			}
			if treeSize%2 != 0 && treeSize > i+1 { // if odd nodes, add one new node
				left := mrkltree.Leafs[len(mrkltree.Leafs)-1].Left
				right := mrkltree.Leafs[len(mrkltree.Leafs)-1].Right
				copyLastNode := NewMerkleNode(left, right, nil)
				mrkltree.Leafs = append(mrkltree.Leafs, copyLastNode)
				treeSize += 1
			}
		}
		i += 2 // increase by two, such that we compare two next nodes in the leafs slice
	}
	return mrkltree
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *Node, data []byte) *Node {
	var hashData [32]byte
	if len(data) == 0 && (right != nil || left != nil) {
		var data []byte
		data = append(data, left.Hash...)
		data = append(data, right.Hash...)
		hashData = sha256.Sum256(data)
	} else {
		hashData = sha256.Sum256(data)
	}
	return &Node{
		Left:  left,
		Right: right,
		Hash:  hashData[:],
	}
}

// MerkleRootHash return the hash of the merkle root node
func (mt *MerkleTree) MerkleRootHash() []byte {
	return mt.RootNode.Hash
}

// MakeMerkleProof returns a list of hashes and indexes required to
// reconstruct the merkle path of a given hash
//
// @param hash represents the hashed data (e.g. transaction ID) stored on
// the leaf node
// @return the merkle proof (list of intermediate hashes), a list of indexes
// indicating the node location in relation with its parent (using the
// constants: leftNode or rightNode), and a possible error.
func (mt *MerkleTree) MakeMerkleProof(hash []byte) ([][]byte, []int64, error) {
	var indexes []int64
	var intermediateHashes [][]byte
	node := findNode(mt.Leafs, hash)
	if node != nil {
		for node != mt.RootNode {
			nodeParent := node.Parent
			if nodeParent.Left == node {
				indexes = append(indexes, rightNode)
				intermediateHashes = append(intermediateHashes, nodeParent.Right.Hash)
			} else {
				indexes = append(indexes, leftNode)
				intermediateHashes = append(intermediateHashes, nodeParent.Left.Hash)
			}
			node = node.Parent
		}
		return intermediateHashes, indexes, nil
	}
	return [][]byte{}, []int64{}, fmt.Errorf("Node %x not found", hash)
}

// VerifyProof verifies that the correct root hash can be retrieved by
// recreating the merkle path for the given hash and merkle proof.
//
// @param rootHash is the hash of the current root of the merkle tree
// @param hash represents the hash of the data (e.g. transaction ID)
// to be verified
// @param mProof is the merkle proof that contains the list of intermediate
// hashes and their location on the tree required to reconstruct
// the merkle path.
func VerifyProof(rootHash []byte, hash []byte, mProof MerkleProof) bool {
	var hashes []byte
	var currentHash [32]byte
	for i, h := range hash {
		currentHash[i] = h
	}
	for i, interHashes := range mProof.proof {
		if mProof.index[i] == leftNode {
			hashes = append(hashes, interHashes...)
			hashes = append(hashes, currentHash[:]...)
			currentHash = sha256.Sum256(hashes)
			hashes = nil
		} else {
			hashes = append(hashes, currentHash[:]...)
			hashes = append(hashes, interHashes...)
			currentHash = sha256.Sum256(hashes)
			hashes = nil
		}
	}
	return bytes.Equal(currentHash[:], rootHash)
}

// firstLevelNodes returns the first level node of Merkle tree
func (mt *MerkleTree) firstLevelNodes(data [][]byte) {
	for _, d := range data {
		newnode := NewMerkleNode(nil, nil, d)
		mt.Leafs = append(mt.Leafs, newnode)
	}
	// if odd txs, add a new node
	if len(mt.Leafs)%2 != 0 {
		copylastNode := NewMerkleNode(nil, nil, data[len(data)-1])
		mt.Leafs = append(mt.Leafs, copylastNode)
	}
}

func (mt *MerkleTree) insertNode(idx int) {
	nxtLevelNode := NewMerkleNode(mt.Leafs[idx-1], mt.Leafs[idx], nil)
	mt.Leafs = append(mt.Leafs, nxtLevelNode)
	mt.Leafs[idx-1].Parent = nxtLevelNode
	mt.Leafs[idx].Parent = nxtLevelNode
}

func findNode(leafs []*Node, hash []byte) *Node {
	for _, n := range leafs {
		if bytes.Equal(n.Hash, hash) {
			return n
		}
	}
	return nil
}
