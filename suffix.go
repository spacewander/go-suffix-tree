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
	// For GetPredecessor and so on. We choice to use more memory(24 bytes per node)
	// over appending keys each time.
	originKey []byte
	value     interface{}
}

type Node struct {
	edges []*Edge
}

func (node *Node) insertEdge(edge *Edge) {
	newEdgeLabelLen := len(edge.label)
	idx := sort.Search(len(node.edges), func(i int) bool {
		return newEdgeLabelLen < len(node.edges[i].label)
	})
	node.edges = append(node.edges, nil)
	copy(node.edges[idx+1:], node.edges[idx:])
	node.edges[idx] = edge
}

func (node *Node) removeEdge(idx int) {
	copy(node.edges[idx:], node.edges[idx+1:])
	node.edges[len(node.edges)-1] = nil
	node.edges = node.edges[:len(node.edges)-1]
}

// Reorder edge which is not shorter than before
func (node *Node) backwardEdge(idx int) {
	edge := node.edges[idx]
	edgeLabelLen := len(edge.label)
	edgesLen := len(node.edges)
	if idx == edgesLen-1 {
		// Still longest, no need to change
		return
	}
	// Get the first edge which's label is longer than this edge...
	i := sort.Search(edgesLen-idx-1, func(j int) bool {
		return edgeLabelLen < len(node.edges[j+idx+1].label)
	})
	// ... and insert before it. (Note that we just add `idx` instead of `idx+1`)
	i += idx
	copy(node.edges[idx:i], node.edges[idx+1:i+1])
	node.edges[i] = edge
}

// Reorder edge which is shorter than before
func (node *Node) forwardEdge(idx int) {
	edge := node.edges[idx]
	edgeLabelLen := len(edge.label)
	i := sort.Search(idx, func(j int) bool {
		return edgeLabelLen < len(node.edges[j].label)
	})
	copy(node.edges[i+1:idx+1], node.edges[i:idx])
	node.edges[i] = edge
}

func (node *Node) insert(originKey []byte, key []byte, value interface{}) (
	oldValue interface{}, ok bool) {

	start := 0
	if len(node.edges) > 0 && len(node.edges[0].label) == 0 {
		// handle empty label as a special case, so the rest of labels don't share
		// common suffix
		if len(key) == 0 {
			leaf, _ := node.edges[0].point.(*Leaf)
			oldValue = leaf.value
			leaf.value = value
			return oldValue, true
		}
		start += 1
	}
	for i := start; i < len(node.edges); i++ {
		edge := node.edges[i]
		gap := suffixDiff(key, edge.label)
		if gap == 0 {
			// CASE 1: key == label
			switch point := edge.point.(type) {
			case *Leaf:
				// Leaf hitted, replace old value
				oldValue = point.value
				point.value = value
				return oldValue, true
			case *Node:
				// Node hitted, insert a leaf under this Node
				return point.insert(originKey, []byte{}, value)
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
								originKey: originKey,
								value:     value,
							},
						},
					},
				}
				edge.point = newNode
				return nil, true
			case *Node:
				// Before: Node - "label" -> Node - "" -> Leaf(Value1)
				// After: Node - "label" - Node - "" -> Leaf(Value1)
				//							|- "s" -> Leaf(Value2)
				// Insert a new Leaf with extra data as label
				return point.insert(originKey, label, value)
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
					originKey: originKey,
					value:     value,
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
			node.forwardEdge(i)
			return nil, true
		}
		// CASE 4: totally mismatch
	}

	leaf := &Leaf{
		originKey: originKey,
		value:     value,
	}
	edge := &Edge{
		label: key,
		point: leaf,
	}
	node.insertEdge(edge)
	return nil, true
}

func (node *Node) get(key []byte) (value interface{}, found bool) {
	edges := node.edges
	start := 0
	if len(edges[0].label) == 0 {
		// handle empty label as a special case, so the rest of labels don't share
		// common suffix
		if len(key) == 0 {
			leaf, _ := edges[0].point.(*Leaf)
			return leaf.value, true
		}
		start += 1
	}

	keyLen := len(key)
	for i := start; i < len(edges); i++ {
		edge := edges[i]
		edgeLabelLen := len(edge.label)
		if keyLen > edgeLabelLen {
			if bytes.Equal(key[len(key)-len(edge.label):], edge.label) {
				key := key[:len(key)-len(edge.label)]
				switch point := edge.point.(type) {
				case *Leaf:
					return nil, false
				case *Node:
					return point.get(key)
				}
			}
		} else if keyLen == edgeLabelLen {
			if bytes.Equal(key, edge.label) {
				switch point := edge.point.(type) {
				case *Leaf:
					return point.value, true
				case *Node:
					return point.get([]byte{})
				}
			}
		} else {
			break
		}
	}

	return nil, false
}

func (node *Node) getPredecessor(key []byte) (matchedKey []byte, value interface{}, found bool) {
	edges := node.edges
	start := 0
	if len(edges[0].label) == 0 {
		// handle empty label as a special case, so the rest of labels don't share
		// common suffix
		if len(key) == 0 {
			leaf, _ := edges[0].point.(*Leaf)
			return leaf.originKey, leaf.value, true
		}
		start += 1
	}

	keyLen := len(key)
	for i := start; i < len(edges); i++ {
		edge := edges[i]
		edgeLabelLen := len(edge.label)
		if keyLen > edgeLabelLen {
			if bytes.Equal(key[len(key)-len(edge.label):], edge.label) {
				key := key[:len(key)-len(edge.label)]
				switch point := edge.point.(type) {
				case *Leaf:
					return point.originKey, point.value, true
				case *Node:
					matchedKey, value, found := point.getPredecessor(key)
					if found {
						return matchedKey, value, found
					}
					// No exact match, fallback to suffix match
				}
			}
		} else if keyLen == edgeLabelLen {
			if bytes.Equal(key, edge.label) {
				switch point := edge.point.(type) {
				case *Leaf:
					return point.originKey, point.value, true
				case *Node:
					matchedKey, value, found := point.getPredecessor([]byte{})
					if found {
						return matchedKey, value, found
					}
				}
			}
		} else {
			break
		}
	}

	if start == 1 {
		leaf, _ := edges[0].point.(*Leaf)
		return leaf.originKey, leaf.value, true
	}

	return nil, nil, false
}

func (node *Node) mergeChildNode(idx int, child *Node) {
	if len(child.edges) == 1 {
		edge := node.edges[idx]
		edge.point = child.edges[0].point
		edge.label = append(child.edges[0].label, edge.label...)
		node.backwardEdge(idx)
	}
	// When child has only one edge, we will remove the child and merge its label,
	// So there is no case that child has no edge.
}

func (node *Node) remove(key []byte) (value interface{}, found bool, childRemoved bool) {
	edges := node.edges
	start := 0
	if len(edges[0].label) == 0 {
		// handle empty label as a special case, so the rest of labels don't share
		// common suffix
		if len(key) == 0 {
			leaf, _ := edges[0].point.(*Leaf)
			value = leaf.value
			node.removeEdge(0)
			return value, true, true
		}
		start += 1
	}

	keyLen := len(key)
	for i := start; i < len(edges); i++ {
		edge := edges[i]
		edgeLabelLen := len(edge.label)
		if keyLen > edgeLabelLen {
			if bytes.Equal(key[len(key)-len(edge.label):], edge.label) {
				key := key[:len(key)-len(edge.label)]
				switch point := edge.point.(type) {
				case *Node:
					value, found, childRemoved = point.remove(key)
					if childRemoved {
						node.mergeChildNode(i, point)
					}
					return value, found, false
				}
			}
		} else if keyLen == edgeLabelLen {
			if bytes.Equal(key, edge.label) {
				switch point := edge.point.(type) {
				case *Leaf:
					value = point.value
					node.removeEdge(i)
					return value, true, true
				case *Node:
					value, found, childRemoved = point.remove([]byte{})
					if childRemoved {
						node.mergeChildNode(i, point)
					}
					return value, found, false
				}
			}
		} else {
			break
		}
	}

	return nil, false, false
}

func (node *Node) walk(suffix []byte, f func(key []byte, value interface{})) {
	for _, edge := range node.edges {
		switch point := edge.point.(type) {
		case *Leaf:
			f(append(edge.label, suffix...), point.value)
		case *Node:
			point.walk(append(edge.label, suffix...), f)
		}
	}
}

func (node *Node) walkNode(suffix [][]byte, f func(labels [][]byte, value interface{})) {
	f(append([][]byte{nil}, suffix...), nil)
	nodes := []*Edge{}
	leaves := []*Edge{}
	for _, edge := range node.edges {
		switch edge.point.(type) {
		case *Leaf:
			leaves = append(leaves, edge)
		case *Node:
			nodes = append(nodes, edge)
		}
	}
	for _, edge := range leaves {
		leaf, _ := edge.point.(*Leaf)
		f(append([][]byte{edge.label}, suffix...), leaf.value)
	}
	for _, edge := range nodes {
		node, _ := edge.point.(*Node)
		node.walkNode(append([][]byte{edge.label}, suffix...), f)
	}
}

type Tree struct {
	root      *Node
	leavesNum int
}

func NewTree() *Tree {
	return &Tree{
		root: &Node{
			edges: []*Edge{},
		},
		leavesNum: 0,
	}
}

// Insert suffix tree with given key and value. Return the previous value and a boolean to
// indicate whether the insertion is successful.
func (tree *Tree) Insert(key []byte, value interface{}) (oldValue interface{}, ok bool) {
	oldValue, ok = tree.root.insert(key, key, value)
	if ok && oldValue == nil {
		tree.leavesNum += 1
	}
	return oldValue, ok
}

// Given a key, Get returns the value itself and a boolean to indicate
// whether the value is found.
func (tree *Tree) Get(key []byte) (value interface{}, found bool) {
	if len(tree.root.edges) == 0 {
		return nil, false
	}
	return tree.root.get(key)
}

// GetPredecessor is mostly like Get.
// It returns the key which is the longest suffix of the given key,
// and the value referred by this key.
// Plus a boolean to indicate whether the key/value, is found.
func (tree *Tree) GetPredecessor(key []byte) (matchedKey []byte, value interface{}, found bool) {
	if len(tree.root.edges) == 0 {
		return nil, nil, false
	}
	return tree.root.getPredecessor(key)
}

// Given a key, Remove returns the value itself and a boolean to indicate
// whethe the value is found. Then the value will be removed.
func (tree *Tree) Remove(key []byte) (oldValue interface{}, found bool) {
	if len(tree.root.edges) == 0 {
		return nil, false
	}
	oldValue, found, _ = tree.root.remove(key)
	if found {
		tree.leavesNum -= 1
	}
	return oldValue, found
}

// Len returns the number of keys.
func (tree *Tree) Len() int {
	return tree.leavesNum
}

// Walk through the tree, call function with key and value. Don't rely on the
// travelling order.
func (tree *Tree) Walk(f func(key []byte, value interface{})) {
	tree.root.walk([]byte{}, f)
}

// This API is for testing/debug
func (tree *Tree) walkNode(f func(labels [][]byte, value interface{})) {
	tree.root.walkNode([][]byte{}, f)
}
