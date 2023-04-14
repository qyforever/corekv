// Copyright 2021 hardcore-os Project Authors
//
// Licensed under the Apache License, Version 2.0 (the "License")
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import "math"

// Filter is an encoded set of []byte keys.
type Filter []byte

// MayContainKey _
func (f Filter) MayContainKey(k []byte) bool {
	return f.MayContain(Hash(k))
}

// MayContain returns whether the filter may contain given key. False positives
// are possible, where it returns true for keys not in the original set.
func (f Filter) MayContain(h uint32) bool {
	//Implement me here!!!
	//在这里实现判断一个数据是否在bloom过滤器中
	//思路大概是经过K个Hash函数计算，判读对应位置是否被标记为1
	if len(f) < 2 {
		return false
	}
	k := f[len(f)-1]
	if k > 30 {
		return true
	}
	nBits := uint32(8 * (len(f) - 1))
	delta := h>>17 | h<<15
	for j := uint8(0); j < k; j++ {
		bitPos := h % nBits
		if f[bitPos/8]&(1<<(bitPos%8)) == 0 {
			return false
		}
		h += delta
	}
	return true
}

// NewFilter returns a new Bloom filter that encodes a set of []byte keys with
// the given number of bits per key, approximately.
//
// A good bitsPerKey value is 10, which yields a filter with ~ 1% false
// positive rate.
func NewFilter(keys []uint32, bitsPerKey int) Filter {
	return Filter(appendFilter(keys, bitsPerKey))
}

// BloomBitsPerKey returns the bits per key required by bloomfilter based on
// the false positive rate.
func BloomBitsPerKey(numEntries int, fp float64) int {
	//Implement me here!!!
	//阅读bloom论文实现，并在这里编写公式
	//传入参数numEntries是bloom中存储的数据个数，fp是false positive假阳性率
	size := -1 * float64(numEntries) * math.Log(fp) / math.Pow(float64(0.69314718056), 2)
	locs := math.Ceil(size / float64(numEntries))
	return int(locs)
}

func appendFilter(keys []uint32, bitsPerKey int) []byte {
	//Implement me here!!!
	//在这里实现将多个Key值放入到bloom过滤器中

	if bitsPerKey < 0 {
		bitsPerKey = 0
	}
	k := uint32(float64(bitsPerKey) * 0.69) // 哈希函数个数
	if k < 1 {
		k = 1
	}
	if k > 30 {
		k = 30
	}
	nbits := len(keys) * int(bitsPerKey) // 元素个数 *
	if nbits < 64 {
		nbits = 64
	}
	nBytes := (nbits + 7) / 8 //上取整
	nBits := nBytes * 8
	filter := make([]byte, nBytes+1)

	for _, h := range keys {
		delta := h>>17 | h<<15
		for j := uint32(0); j < k; j++ {
			bitPos := h % uint32(nBits)
			filter[bitPos/8] |= 1 << (bitPos % 8)
			h += delta
		}
	}
	filter[nBytes] = uint8(k)
	return filter
}

// Hash implements a hashing algorithm similar to the Murmur hash.

// 32-bit mixing function.

func Hash(b []byte) uint32 {
	//Implement me here!!!
	//在这里实现高效的HashFunction

	// Initialize the hash to a 'random' value
	const (
		M    = 0xc6a4a793
		seed = 0xbc9f1d34
	)
	h := uint32(seed) ^ uint32(len(b))*M

	// Mix 4 bytes at a time into the hash
	for ; len(b) >= 4; b = b[4:] {
		h += uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
		h *= M
		h ^= h >> 16
	}

	// Handle the last few bytes of the input array
	switch len(b) {
	case 3:
		h += uint32(b[2]) << 16
		fallthrough
	case 2:
		h += uint32(b[1]) << 8
		fallthrough
	case 1:
		h += uint32(b[0])
		h *= M
		h ^= h >> 24
	}
	// Do a few final mixes of the hash to ensure the last few bytes are well incorporated
	return h
}
