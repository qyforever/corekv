package utils

import (
	"github.com/pkg/errors"
	"log"
	"sync/atomic"
	"unsafe"
)

type Arena struct {
	n          uint32 //offset 当前Arena已分配出去的内存大小
	shouldGrow bool
	buf        []byte // Arena申请的内存空间
}

const (
	MaxNodeSize = int(unsafe.Sizeof(Element{}))
	offsetSize  = int(unsafe.Sizeof(uint32(0))) // 分配一个32位无符号整数的内存大小
	nodeAlign   = int(unsafe.Sizeof(uint64(0))) - 1
)

func newArena(n int64) *Arena {
	out := &Arena{
		n:   1,
		buf: make([]byte, n),
	}
	return out
}

func (s *Arena) allocate(sz uint32) uint32 {
	//implement me here！！！
	// 在 arena 中分配指定大小的内存空间
	offset := atomic.AddUint32(&s.n, sz)
	if !s.shouldGrow {
		AssertTrue(int(offset) <= len(s.buf))
		return offset - sz
	}
	if int(offset) > len(s.buf)-MaxNodeSize {
		growBy := uint32(len(s.buf))
		if growBy > 1<<30 {
			growBy = 1 << 30
		}
		if growBy < sz {
			growBy = sz
		}

		newBuf := make([]byte, len(s.buf)+int(growBy))
		AssertTrue(len(s.buf) == copy(newBuf, s.buf))
		s.buf = newBuf
	}
	return offset - sz // 返回起始地址
}

// 在arena里开辟一块空间，用以存放sl中的节点
// 返回值为在arena中的offset
func (s *Arena) putNode(height int) uint32 {
	//implement me here！！！
	// 这里的 node 要保存 value 、key 和 next 指针值
	// 所以要计算清楚需要申请多大的内存空间
	unusedSize := (defaultMaxLevel - height) * offsetSize
	l := uint32(MaxNodeSize - unusedSize + nodeAlign)
	n := s.allocate(l)

	m := (n + uint32(nodeAlign)) &^ uint32(nodeAlign) //将运算符左边数据相异的位保留，相同位清零
	return m
	// levels 里面需要的大小

}

func (s *Arena) putVal(v ValueStruct) uint32 {
	//implement me here！！！
	//将 Value 值存储到 arena 当中
	// 并且将指针返回，返回的指针值应被存储在 Node 节点中
	l := uint32(v.EncodedSize())
	offset := s.allocate(l)
	v.EncodeValue(s.buf[offset:])

	return offset
}

func (s *Arena) putKey(key []byte) uint32 {
	//implement me here！！！
	//将  Key 值存储到 arena 当中
	// 并且将指针返回，返回的指针值应被存储在 Node 节点中

	l := len(key)
	offset := s.allocate(uint32(l))

	bufset := s.buf[offset : offset+uint32(l)]

	AssertTrue(l == copy(bufset, key))
	return offset
}

func (s *Arena) getElement(offset uint32) *Element {
	if offset == 0 {
		return nil
	}

	return (*Element)(unsafe.Pointer(&s.buf[offset]))
}

func (s *Arena) getKey(offset uint32, size uint16) []byte {
	return s.buf[offset : offset+uint32(size)]
}

func (s *Arena) getVal(offset uint32, size uint32) (v ValueStruct) {
	v.DecodeValue(s.buf[offset : offset+size])
	return
}

// 用element在内存中的地址 - arena首字节的内存地址，得到在arena中的偏移量
func (s *Arena) getElementOffset(nd *Element) uint32 {
	//implement me here！！！
	//获取某个节点，在 arena 当中的偏移量
	if nd == nil {
		return 0
	}

	return uint32(uintptr(unsafe.Pointer(nd)) - uintptr(unsafe.Pointer(&s.buf[0])))

}

func (e *Element) getNextOffset(h int) uint32 {
	//implement me here！！！
	// 这个方法用来计算节点在h 层数下的 next 节点
	return atomic.LoadUint32(&e.levels[h])
}

func (s *Arena) Size() int64 {
	return int64(atomic.LoadUint32(&s.n))
}

func AssertTrue(b bool) {
	if !b {
		log.Fatalf("%+v", errors.Errorf("Assert failed"))
	}
}
