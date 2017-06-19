package suffix

import (
	"fmt"
)

func ExampleTree_Insert() {
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

func ExampleTree_Get() {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	value, found := tree.Get([]byte("sth"))
	if found {
		fmt.Println(value.(string))
	}
	// Output: sth
}

func ExampleTree_LongestSuffix() {
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

func ExampleTree_Remove() {
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

func ExampleTree_Walk() {
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

func ExampleTree_WalkSuffix() {
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
