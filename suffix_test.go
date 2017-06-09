package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	RunAlhoc = flag.Bool("alhoc", false, "Run alhoc testing")
)

// go test -args -alhoc to enable alhoc testing
func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
}

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

// GetPredecessor is mostly like Get, but please notice their slight differnces.
func assertGetPredecessor(t *testing.T, tree *Tree, expectedValue string,
	expectedFound bool) {

	_, value, found := tree.GetPredecessor([]byte(expectedValue))
	assert.Equal(t, expectedFound, found, "expected %s, got nothing", expectedValue)
	if expectedFound && value != nil {
		assert.Equal(t, expectedValue, value.(string))
	}
}

func assertGetPredecessorCheckKey(t *testing.T, tree *Tree, expectedValue string,
	expectedFound bool, expectedKey string) {

	key, value, found := tree.GetPredecessor([]byte(expectedValue))
	assert.Equal(t, expectedFound, found, "expected %s, got nothing", expectedValue)
	if expectedFound && value != nil {
		assert.Equal(t, expectedKey, string(key))
		assert.Equal(t, expectedKey, value.(string))
	}
}

func TestGetPredecessor_EmptyTree(t *testing.T) {
	tree := NewTree()
	assertGetPredecessor(t, tree, "banana", false)
}

func TestGetPredecessor_Base(t *testing.T) {
	tree := NewTree()
	tree.Insert([]byte("sth"), "sth")
	assertGetPredecessor(t, tree, "th", false)
	assertGetPredecessor(t, tree, "else", false)
	assertGetPredecessorCheckKey(t, tree, "sth", true, "sth")
	assertGetPredecessorCheckKey(t, tree, "any sth", true, "sth")

	tree.Insert([]byte("else sth"), "else sth")
	assertGetPredecessor(t, tree, "sth", true)
	assertGetPredecessorCheckKey(t, tree, "lse sth", true, "sth")
	assertGetPredecessorCheckKey(t, tree, "any else sth", true, "else sth")

	tree.Insert([]byte("any sth"), "tenth")
	assertGetPredecessor(t, tree, "th", false)
	assertGetPredecessor(t, tree, "fourth", false)
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

func dumpTestData(wordRef map[string]bool, tree *Tree, ops []string, errMsg string) {
	tmpfile, _ := ioutil.TempFile("", "suffix_test_")
	defer tmpfile.Close()
	for _, op := range ops {
		println(op)
		tmpfile.Write(append([]byte(op), []byte("\n")...))
	}
	println("\nWord status:")
	for word, existed := range wordRef {
		if existed {
			println(word, "existed")
		} else {
			println(word, "removed")
		}
	}
	println("\nTree nodes:")
	tree.walkNode(func(labels [][]byte, value interface{}) {
		if labels[0] == nil {
			return
		}
		suffixes := []string{}
		for _, l := range labels {
			suffixes = append(suffixes, string(l))
		}
		println(strings.Join(suffixes, ":"))
	})
	println(errMsg)
	println("Also dump operation records to", tmpfile.Name())
}

func checkLabelOrder(tree *Tree) (string, bool) {
	var preLabelLen int
	var msg string
	inOrder := true
	tree.walkNode(func(labels [][]byte, value interface{}) {
		if labels[0] != nil {
			if len(labels[0]) < preLabelLen {
				msg = fmt.Sprintf("expect label len not shorter than %d, actual len(%s) is %d",
					preLabelLen, string(labels[0]), len(labels[0]))
				inOrder = false
			}
			preLabelLen = len(labels[0])
		} else {
			preLabelLen = 0
		}
	})
	return msg, inOrder
}

func TestAlhoc(t *testing.T) {
	if !*RunAlhoc {
		t.SkipNow()
	}
	println(`
Start alhoc test.
Repeat below steps in 30 seconds.
1. Generate 256 random words, and insert them into a new Tree.
2. Perform 2048 random operations with pre-generated 256 words.
3. Dump the generated test data once failed.
`)
	OpNum := 2048
	WordNum := 256
	// Put some variable definitions outside for loop,
	// so that we could refer it in test dump.
	wordRef := map[string]bool{}
	randomWords := []string{}
	ops := []string{}
	rand.Seed(time.Now().UnixNano())
	letters := []byte("abcdefghijklmnopqrstuvwxyz")
	mismatchLetters := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	testTurns := 0
	testEnd := time.NewTimer(30 * time.Second)
	var errMsg string
	var tree *Tree
	defer func() {
		if r := recover(); r != nil {
			errMsg = fmt.Sprintf("Panic happened: %v", r)
			dumpTestData(wordRef, tree, ops, errMsg)
			t.FailNow()
		}
	}()

	for {
		select {
		case <-testEnd.C:
			println(testTurns, "turns of alhoc tests pass.")
			println("Alhoc test is finished.")
			return
		default:
		}
		tree = NewTree()
		wordRef = map[string]bool{}
		randomWords = []string{}
		ops = []string{}
		for wordCount := 0; wordCount < WordNum; {
			b := make([]byte, rand.Intn(12))
			for i := range b {
				b[i] = letters[rand.Intn(len(letters))]
			}
			bs := string(b)
			if _, ok := wordRef[bs]; !ok {
				wordRef[bs] = true
				wordCount += 1
			}
			ops = append(ops, "Insert\t"+bs)
			tree.Insert(b, bs)
			value, found := tree.Get(b)
			if !found {
				errMsg = fmt.Sprintf("expect get %v after insertion, actual not found", bs)
			} else {
				if value.(string) != bs {
					errMsg = fmt.Sprintf("expect insert %v, actual %v", bs, value.(string))
					goto failed
				}
			}
			msg, inOrder := checkLabelOrder(tree)
			if !inOrder {
				errMsg = msg
				goto failed
			}
		}
		for s := range wordRef {
			randomWords = append(randomWords, s)
		}
		for i := 0; i < OpNum; i++ {
			if OpNum%256 == 0 {
				existedNum := 0
				for _, existed := range wordRef {
					if existed {
						existedNum += 1
					}
				}
				if tree.Len() != existedNum {
					errMsg = fmt.Sprintf("expect len %v, actual %v", existedNum, tree.Len())
					goto failed
				}
			}
			word := randomWords[rand.Intn(WordNum)]
			switch rand.Intn(4) {
			case 0:
				existed := wordRef[word]
				ops = append(ops, "Get\t"+word)
				value, found := tree.Get([]byte(word))
				if found {
					if !existed {
						errMsg = fmt.Sprintf("expect not found %v, actual found", word)
						goto failed
					}
					if value.(string) != word {
						errMsg = fmt.Sprintf("expect get %v, actual %v", word, value.(string))
						goto failed
					}
				} else if existed {
					errMsg = fmt.Sprintf("expect found %v, actual not found", word)
					goto failed
				}
			case 1:
				existed := wordRef[word]
				ops = append(ops, "Insert\t"+word)
				value, _ := tree.Insert([]byte(word), word)
				if existed {
					if value.(string) != word {
						errMsg = fmt.Sprintf("expect insert %v, actual %v", word, value.(string))
						goto failed
					}
				}
				wordRef[word] = true
				_, found := tree.Get([]byte(word))
				if !found {
					errMsg = fmt.Sprintf("expect get %v after insertion, actual not found", word)
				}
				msg, inOrder := checkLabelOrder(tree)
				if !inOrder {
					errMsg = msg
					goto failed
				}
			case 2:
				existed := wordRef[word]
				ops = append(ops, "Remove\t"+word)
				value, found := tree.Remove([]byte(word))
				wordRef[word] = false
				if found {
					if !existed {
						errMsg = fmt.Sprintf("expect not found %v in removal, actual found", word)
						goto failed
					}
					if value.(string) != word {
						errMsg = fmt.Sprintf("expect remove %v, actual %v", word, value.(string))
						goto failed
					}
					_, found = tree.Get([]byte(word))
					if found {
						errMsg = fmt.Sprintf("expect %v not found after removal, actual found", word)
					}
					msg, inOrder := checkLabelOrder(tree)
					if !inOrder {
						errMsg = msg
						goto failed
					}
				} else if existed {
					errMsg = fmt.Sprintf("expect found %v in removal, actual not found", word)
					goto failed
				}
			case 3:
				existed := wordRef[word]
				mismatchLabel := make([]byte, rand.Intn(3))
				for i := range mismatchLabel {
					mismatchLabel[i] = mismatchLetters[rand.Intn(len(mismatchLetters))]
				}
				suffix := string(mismatchLabel) + word
				ops = append(ops, "GetPredecessor\t"+suffix)
				key, value, found := tree.GetPredecessor([]byte(suffix))
				if existed {
					if !found {
						errMsg = fmt.Sprintf("expect getPredecessor found %v with %v, actual not found",
							word, suffix)
						goto failed
					}
					if value.(string) != string(key) {
						errMsg = fmt.Sprintf(
							"expect getPredecessor %v, actual %v", string(key), value.(string))
						goto failed
					}
				} else {
					if found {
						if !strings.HasSuffix(suffix, string(key)) {
							errMsg = fmt.Sprintf(
								"expect getPredecessor found suffix matched %v, actual found %v",
								suffix, string(key))
							goto failed
						}
					}
				}
			}
		}
		testTurns++
	}
failed:
	dumpTestData(wordRef, tree, ops, errMsg)
	t.FailNow()
}
