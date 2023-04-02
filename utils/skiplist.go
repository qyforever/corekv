package utils

import (
	"bytes"
	"github.com/hardcore-os/corekv/utils/codec"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultMaxLevel = 48
)

type SkipList struct {
	header *Element

	rand *rand.Rand

	maxLevel int
	length   int
	lock     sync.RWMutex
	size     int64
}

func NewSkipList() *SkipList {
	//implement me here!!!
	seed := rand.NewSource(time.Now().UnixNano())

	return &SkipList{
		header: &Element{
			levels: make([]*Element, defaultMaxLevel),
			entry:  nil,
			score:  0,
		},
		rand:     rand.New(seed),
		maxLevel: defaultMaxLevel,
		length:   0,
	}
}

type Element struct {
	levels []*Element
	entry  *codec.Entry
	score  float64
}

func newElement(score float64, entry *codec.Entry, level int) *Element {
	return &Element{
		levels: make([]*Element, level),
		entry:  entry,
		score:  score,
	}
}

func (elem *Element) Entry() *codec.Entry {
	return elem.entry
}

func (list *SkipList) Add(data *codec.Entry) error {
	// //implement me here!!!
	list.lock.Lock()
	defer list.lock.Unlock()
	score := list.calcScore(data.Key)
	var e *Element

	max := len(list.header.levels)
	prevElem := list.header
	var prevElemlist [defaultMaxLevel]*Element

	for i := max - 1; i >= 0; { //i层
		prevElemlist[i] = prevElem

		for next := prevElem.levels[i]; next != nil; next = prevElem.levels[i] {
			//在每一层执行查找，当下一个与元素值大于
			if cmp := list.compare(score, data.Key, next); cmp <= 0 {
				if cmp == 0 {
					e = next
					e.entry = data
					list.size += e.Entry().Size() - data.Size()
					return nil
				}
				break
			}
			prevElem = next
			prevElemlist[i] = prevElem
		}
		topLevel := prevElem.levels[i]

		//to skip same prevHeader's next and fill next elem into temp element
		for i--; i >= 0 && prevElem.levels[i] == topLevel; i-- {
			prevElemlist[i] = prevElem
		}
	}
	level := list.randLevel()

	e = newElement(score, data, level)
	for i := 0; i < level; i++ {
		e.levels[i] = prevElemlist[i].levels[i]
		prevElemlist[i].levels[i] = e
	}
	list.size += data.Size()
	list.length++
	return nil
}

func (list *SkipList) Search(key []byte) (e *codec.Entry) {
	//implement me here!!!
	//pre := list.header
	//cur :=
	list.lock.RLock()
	defer list.lock.RUnlock()
	if list.length == 0 {
		return nil
	}
	score := list.calcScore(key)
	prevElem := list.header
	i := len(list.header.levels) - 1

	for i >= 0 {
		for next := prevElem.levels[i]; next != nil; next = prevElem.levels[i] {
			if cmp := list.compare(score, key, next); cmp <= 0 {
				if cmp == 0 {
					return next.Entry()
				}
				break
			}
			prevElem = next
		}
		topLevel := prevElem.levels[i]

		for i--; i >= 0 && prevElem.levels[i] == topLevel; i-- {

		}
	}

	return nil
}

func (list *SkipList) Close() error {
	return nil
}

func (list *SkipList) calcScore(key []byte) (score float64) {
	var hash uint64
	l := len(key)

	if l > 8 {
		l = 8
	}

	for i := 0; i < l; i++ {
		shift := uint(64 - 8 - i*8)
		hash |= uint64(key[i]) << shift
	}

	score = float64(hash)
	return score
}

func (list *SkipList) compare(score float64, key []byte, next *Element) int {
	//implement me here!!!
	if score == next.score {
		return bytes.Compare(key, next.entry.Key)
	}
	if score < next.score {
		return -1
	} else {
		return 1
	}
}

func (list *SkipList) randLevel() int {
	//implement me here!!!
	if list.maxLevel <= 1 {
		return 1
	}
	i := 1
	for ; i < list.maxLevel; i++ {
		if RandN(1000)%2 == 0 {
			return i
		}
	}
	return 0
}

func (list *SkipList) Size() int64 {
	//implement me here!!!
	return list.size

}
