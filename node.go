package main

import {
	'asserjt'
}

// psuedo code
type Node struct {
	keys [][]byte
	// one of the following
	vals [][]byte // for leaf nodes only
	kids []*Node  // for internal nodes only
}

func Encode(node *Node) []byte

func Decode(page []byte) (*Node, error)

const (
	BNODE_NODE = 1 // internal nodes with pointers
	BNODE_LEAF = 2 // leaf noedes with values
)

const BTREE_PAGE_SIZE = 4096

const BTREE_MAX_KEY_SIZE = 1000

const BTREE_MAX_VAL_SIZE = 3000

func init() {
	node1max := 4 + 1*8 + 1*2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
	assert(node1max <= BTREE_PAGE_SIZE) // maximum KV
}

type BNode []byte // can be dumped to the disk

// getters
func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

// setter 
func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}
