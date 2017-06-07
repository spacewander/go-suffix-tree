package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getFixtures() ([]string, *Tree) {
	tree := NewTree()
	lists := []string{
		"edible", "presentable", "abominable", "credible",
		"picturesque", "statuesque", "nothing", "something", "thing", "nonsense",
		"random word", "word", "table", "unbelievable", "believable", "sense",
	}
	for _, s := range lists {
		tree.Insert([]byte(s), s)
	}
	return lists, tree
}

func TestInsertReturn(t *testing.T) {
	tree := NewTree()
	oldValue, ok := tree.Insert([]byte("sth"), "sth")
	assert.True(t, ok)
	assert.Nil(t, oldValue)

	oldValue, ok = tree.Insert([]byte("sth"), "else")
	assert.True(t, ok)
	assert.Equal(t, "sth", oldValue.(string))
}

func assertGet(t *testing.T, tree *Tree, expectedValue string, expectedFound bool) {
	value, found := tree.Get([]byte(expectedValue))
	assert.Equal(t, expectedFound, found, "expected %s, got nothing", expectedValue)
	if expectedFound && value != nil {
		assert.Equal(t, expectedValue, value.(string))
	}
}

func TestGet_EmptyTree(t *testing.T) {
	tree := NewTree()
	assertGet(t, tree, "sth", false)
}

func TestGet_Base(t *testing.T) {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	assertGet(t, tree, "sth", true)
	assertGet(t, tree, "else", false)
	assertGet(t, tree, "any sth", false)

	tree = NewTree()
	tree.Insert([]byte("sth else"), "sth else")
	tree.Insert([]byte("else"), "else")
	assertGet(t, tree, "else", true)
	assertGet(t, tree, "sth else", true)

	tree = NewTree()
	tree.Insert([]byte("else"), "else")
	tree.Insert([]byte("sth else"), "sth else")
	tree.Insert([]byte("any sth else"), "any sth else")
	tree.Insert([]byte("anything sth else"), "anything sth else")
	tree.Insert([]byte("at any sth else"), "at any sth else")
	assertGet(t, tree, "else", true)
	assertGet(t, tree, "sth else", true)
	assertGet(t, tree, "any sth else", true)
}

func TestGet_Random(t *testing.T) {
	lists, tree := getFixtures()
	for _, s := range lists {
		assertGet(t, tree, s, true)
	}
}

func TestRemove_EmptyTree(t *testing.T) {
	tree := NewTree()
	_, found := tree.Remove([]byte("anything"))
	assert.False(t, found)
}

func TestRemove_Base(t *testing.T) {
	tree := NewTree()
	tree.Insert([]byte("else"), "else")
	_, found := tree.Remove([]byte("lse"))
	assert.False(t, found)
	_, found = tree.Remove([]byte("anything"))
	assert.False(t, found)

	tree.Insert([]byte("sth else"), "sth else")
	oldValue, found := tree.Remove([]byte("sth else"))
	assert.True(t, found)
	assert.Equal(t, "sth else", oldValue.(string))
	assertGet(t, tree, "sth else", false)
	assertGet(t, tree, "else", true)

	tree.Remove([]byte("else"))
	assertGet(t, tree, "else", false)

	tree = NewTree()
	tree.Insert([]byte("sth else"), "sth else")
	tree.Insert([]byte("else"), "else")
	_, found = tree.Remove([]byte("else"))
	assert.True(t, found)
	assertGet(t, tree, "else", false)
	assertGet(t, tree, "sth else", true)
}

func TestRemove_Random(t *testing.T) {
	lists, tree := getFixtures()
	for _, s := range lists {
		t.Log("Try to remove ", s)
		assertGet(t, tree, s, true)
		oldValue, found := tree.Remove([]byte(s))
		assert.True(t, found)
		assert.Equal(t, s, oldValue.(string))
		assertGet(t, tree, s, false)
	}
}

func TestWalk(t *testing.T) {
	lists, tree := getFixtures()
	result := map[string]string{}
	tree.Walk(func(key []byte, value interface{}) {
		result[string(key)] = value.(string)
	})

	for _, s := range lists {
		assert.Equal(t, s, result[s])
	}
}
