type Node struct {
	keys [][]byte
	vals [][]byte
	kids []*Node
}


func Encode(node *Node) []byte

func Decode(page []byte) (*Node, error)


