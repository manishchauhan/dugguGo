package tree

import "fmt"

// Node represents a node in the binary search tree
type Node struct {
	data  int
	left  *Node
	right *Node
}

// Tree represents a binary search tree
type Tree struct {
	root *Node
}

// Insert adds a new node with the given data to the binary search tree
func (t *Tree) Insert(data int) {
	newNode := &Node{data: data}
	if t.root == nil {
		t.root = newNode
	} else {
		t.insertRecursively(t.root, newNode)
	}
}

// Helper function to insert a node recursively
func (t *Tree) insertRecursively(current, newNode *Node) {
	if newNode.data < current.data {
		if current.left == nil {
			current.left = newNode
		} else {
			t.insertRecursively(current.left, newNode)
		}
	} else {
		if current.right == nil {
			current.right = newNode
		} else {
			t.insertRecursively(current.right, newNode)
		}
	}
}

// InOrderTraversal prints the elements of the binary search tree in ascending order
func (t *Tree) InOrderTraversal() {
	t.inOrderRecursively(t.root)
	fmt.Println()
}

// Helper function for in-order traversal
func (t *Tree) inOrderRecursively(current *Node) {
	if current != nil {
		t.inOrderRecursively(current.left)
		fmt.Printf("%d ", current.data)
		t.inOrderRecursively(current.right)
	}
}
