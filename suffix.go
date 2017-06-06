package main

import (
	"bytes"
	"sort"
)

// Return
// the first index of the mismatch byte (from right to left, starts from 1)
// len(left)+1 if left byte sequence is shorter than right one
// 0 if two byte sequences are equal
// -len(right)-1 if left byte sequence is longer than right one
func suffixDiff(left, right []byte) int {
	leftLen := len(left)
	rightLen := len(right)
	minLen := leftLen
	if minLen > rightLen {
		minLen = rightLen
	}
	for i := 1; i <= minLen; i++ {
		if left[leftLen-i] != right[rightLen-i] {
			return i
		}
	}
	if leftLen < rightLen {
		return leftLen + 1
	} else if leftLen == rightLen {
		return 0
	}
	return -rightLen - 1
}

type Edge struct {
	label []byte
	// Could be either Node or Leaf
	point interface{}
}

type Leaf struct {
	value interface{}
}

type Node struct {
	edges []*Edge
}

func (node *Node) insert(key []byte, value interface{}) (ok bool, oldValue interface{}) {
	for i, edge := range node.edges {
		gap := suffixDiff(key, edge.label)
		if gap == 0 {
			// CASE 1: key == label
			switch point := edge.point.(type) {
			case *Leaf:
				// Leaf hitted, replace old value
				oldValue = point.value
				point.value = value
				return true, oldValue
			case *Node:
				// Node hitted, insert a leaf under this Node
				return point.insert([]byte{}, value)
			}
		} else if gap < 0 {
			// CASE 2: key > label
			gap = -gap
			label := key[:len(key)-gap+1]
			switch point := edge.point.(type) {
			case *Leaf:
				// Before: Node - "label" -> Leaf(Value1)
				// After: Node - "label" - Node - "" -> Leaf(Value1)
				//							|- "s" -> Leaf(Value2)
				// Create new Node, move old Leaf under new Node, and then
				//	insert a new Leaf
				newNode := &Node{
					edges: []*Edge{
						&Edge{
							label: []byte{},
							point: point,
						},
						&Edge{
							label: label,
							point: &Leaf{
								value: value,
							},
						},
					},
				}
				edge.point = newNode
				return true, nil
			case *Node:
				// Before: Node - "label" -> Node - "" -> Leaf(Value1)
				// After: Node - "label" - Node - "" -> Leaf(Value1)
				//							|- "s" -> Leaf(Value2)
				// Insert a new Leaf with extra data as label
				return point.insert(label, value)
			}
		} else if gap > 1 {
			// CASE 3: mismatch(key, label) after first letter or key < label
			// Before: Node - "labels" -> Node/Leaf(Value1)
			// After: Node - "label" - Node - "s" -> Node/Leaf(Value1)
			//						    |- "" -> Leaf(Value2)
			// Before: Node - "label" -> Node/Leaf(Value1)
			// After: Node - "lab" - Node - "el" -> Node/Leaf(Value1)
			//							|- "or" -> Leaf(Value2)
			newEdge := &Edge{
				label: edge.label[:len(edge.label)-gap+1],
				point: edge.point,
			}
			keyEdge := &Edge{
				label: key[:len(key)-gap+1],
				point: &Leaf{
					value: value,
				},
			}
			newNode := &Node{
				edges: make([]*Edge, 2),
			}
			if len(newEdge.label) < len(keyEdge.label) {
				newNode.edges[0], newNode.edges[1] = newEdge, keyEdge
			} else {
				newNode.edges[0], newNode.edges[1] = keyEdge, newEdge
			}
			edge.point = newNode
			edge.label = edge.label[len(edge.label)-gap+1:]
			idx := sort.Search(i, func(j int) bool {
				return len(edge.label) < len(node.edges[j].label)
			})
			copy(node.edges[idx+1:i+1], node.edges[idx:i])
			node.edges[idx] = edge
			return true, nil
		}
		// CASE 4: totally mismatch
	}

	leaf := &Leaf{
		value: value,
	}
	edge := &Edge{
		label: key,
		point: leaf,
	}
	idx := sort.Search(len(node.edges), func(i int) bool {
		return len(key) < len(node.edges[i].label)
	})
	node.edges = append(node.edges, nil)
	copy(node.edges[idx+1:], node.edges[idx:])
	node.edges[idx] = edge
	return true, nil
}

func (node *Node) get(key []byte) (found bool, value interface{}) {
	edges := node.edges
	if len(edges[0].label) == 0 {
		// handle empty label as a special case, so the rest of labels don't share
		// common suffix
		if len(key) == 0 {
			leaf, _ := edges[0].point.(Leaf)
			return true, leaf.value
		}
		edges = edges[1:]
	}

	keyLen := len(key)
	for i := 0; i < len(edges); i++ {
		edge := edges[i]
		edgeLabelLen := len(edge.label)
		if keyLen > edgeLabelLen {
			if bytes.Equal(key[len(key)-len(edge.label):], edge.label) {
				key := key[:len(key)-len(edge.label)]
				switch point := edge.point.(type) {
				case *Node:
					return point.get(key)
				}
			}
		} else if bytes.Equal(key, edge.label) {
			switch point := edge.point.(type) {
			case *Leaf:
				return true, point.value
			case *Node:
				return point.get([]byte{})
			}
		} else {
			break
		}
	}

	return false, nil
}

func (node *Node) Walk(suffix string, f func(key string, value interface{})) {
	for _, edge := range node.edges {
		switch point := edge.point.(type) {
		case *Leaf:
			f(string(edge.label)+suffix, point.value)
		case *Node:
			point.Walk(string(edge.label)+suffix, f)
		}
	}
}

type Tree struct {
	root *Node
}

func NewTree() *Tree {
	return &Tree{
		root: &Node{
			edges: []*Edge{},
		},
	}
}

func (tree *Tree) Insert(key []byte, value interface{}) (ok bool, oldValue interface{}) {
	return tree.root.insert(key, value)
}

func (tree *Tree) Get(key []byte) (found bool, value interface{}) {
	if len(tree.root.edges) == 0 {
		return false, nil
	}
	return tree.root.get(key)
}

func (tree *Tree) Walk(f func(key string, value interface{})) {
	tree.root.Walk("", f)
}
