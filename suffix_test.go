package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertReturn(t *testing.T) {
	tree := NewTree()
	ok, oldValue := tree.Insert([]byte("sth"), "sth")
	assert.True(t, ok)
	assert.Nil(t, oldValue)

	ok, oldValue = tree.Insert([]byte("sth"), "else")
	assert.True(t, ok)
	assert.Equal(t, "sth", oldValue.(string))
}

func assertGet(t *testing.T, tree *Tree, expectedValue string, expectedFound bool) {
	found, value := tree.Get([]byte(expectedValue))
	assert.Equal(t, expectedFound, found, "expected ", expectedValue, "got nothing")
	if expectedFound && value != nil {
		assert.Equal(t, expectedValue, value.(string))
	}
}

func TestGet_EmptyTree(t *testing.T) {
	tree := NewTree()
	assertGet(t, tree, "sth", false)
}

func TestGet(t *testing.T) {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	assertGet(t, tree, "sth", true)
	assertGet(t, tree, "else", false)
	assertGet(t, tree, "any sth", false)
}

func TestGet_AncestorNodes(t *testing.T) {
	tree := NewTree()
	tree.Insert([]byte("else"), "else")
	tree.Insert([]byte("sth else"), "sth else")
	assertGet(t, tree, "else", true)
	assertGet(t, tree, "sth else", true)
	assertGet(t, tree, "any sth else", false)

	tree = NewTree()
	tree.Insert([]byte("sth else"), "sth else")
	tree.Insert([]byte("else"), "else")
	assertGet(t, tree, "else", true)
	assertGet(t, tree, "sth else", true)
}

func TestWalk(t *testing.T) {
	tree := NewTree()
	lists := []string{
		"edible", "presentable", "abominable", "credible",
		"picturesque", "statuesque", "nothing", "something", "thing", "nonsense",
	}
	for _, s := range lists {
		tree.Insert([]byte(s), s)
	}

	result := map[string]string{}
	tree.Walk(func(key []byte, value interface{}) {
		result[string(key)] = value.(string)
	})

	for _, s := range lists {
		assert.Equal(t, s, result[s])
	}
}
