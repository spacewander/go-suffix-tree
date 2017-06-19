package suffix

import (
	"fmt"
)

func ExampleInsert() {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	oldValue, ok := tree.Insert([]byte("sth"), "else")
	// Always ok
	if ok {
		fmt.Println(oldValue.(string))
	}
	value, found := tree.Get([]byte("sth"))
	if found {
		fmt.Println(value.(string))
	}
	// Output:
	// sth
	// else
}

func ExampleGet() {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	value, found := tree.Get([]byte("sth"))
	if found {
		fmt.Println(value.(string))
	}
	// Output: sth
}

func ExampleLongestSuffix() {
	tree := NewTree()
	tree.Insert([]byte("table"), "table")
	tree.Insert([]byte("able"), "able")
	tree.Insert([]byte("present"), "present")
	key, value, found := tree.LongestSuffix([]byte("presentable"))
	if found {
		fmt.Println("Matched key:", string(key))
		fmt.Println(value.(string))
	}
	// Output:
	// Matched key: table
	// table
}

func ExampleRemove() {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	oldValue, found := tree.Remove([]byte("sth"))
	if found {
		fmt.Println(oldValue.(string))
	}
	_, found = tree.Get([]byte("sth"))
	if !found {
		fmt.Println("Already removed")
	}
	// Output:
	// sth
	// Already removed
}

func ExampleWalk() {
	tree := NewTree()
	tree.Insert([]byte("able"), 1)
	tree.Insert([]byte("table"), 2)
	tree.Insert([]byte("presentable"), 3)
	tree.Walk(func(key []byte, _ interface{}) (stop bool) {
		fmt.Println(string(key))
		return false
	})
	fmt.Println("Walk and stop in the middle:")
	tree.Walk(func(key []byte, _ interface{}) (stop bool) {
		fmt.Println(string(key))
		return string(key) == "able"
	})
	// Output:
	// able
	// table
	// presentable
	// Walk and stop in the middle:
	// able
}

func ExampleWalkSuffix() {
	tree := NewTree()
	tree.Insert([]byte("able"), 1)
	tree.Insert([]byte("table"), 2)
	tree.Insert([]byte("presentable"), 3)
	tree.Insert([]byte("present"), 4)
	tree.WalkSuffix([]byte("table"), func(key []byte, _ interface{}) (stop bool) {
		fmt.Println(string(key))
		return false
	})
	fmt.Println("Walk and stop in the middle:")
	tree.WalkSuffix([]byte("table"), func(key []byte, _ interface{}) (stop bool) {
		fmt.Println(string(key))
		return string(key) == "table"
	})
	// Output:
	// table
	// presentable
	// Walk and stop in the middle:
	// table
}
