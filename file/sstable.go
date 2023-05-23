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

package file

// TODO LAB 在这里实现 sst 文件操作
import (
	"github.com/golang/protobuf/proto"
	"github.com/hardcore-os/corekv/utils"
	"github.com/hardcore-os/corekv/utils/codec"
	"github.com/hardcore-os/corekv/utils/codec/pb"
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
)

type SSTable struct {
	lock           *sync.RWMutex
	f              *MmapFile
	maxKey         []byte
	minKey         []byte
	idxTables      *pb.TableIndex
	hasBloomFilter bool
	idxLen         int
	idxStart       int
	fid            uint64
}

// 打开一个sstable文件
func OpenSSTable(opt *Options) *SSTable {
	omf, err := OpenMmapFile(opt.FileName, os.O_CREATE|os.O_RDWR, opt.MaxSz)
	utils.Err(err)
	return &SSTable{f: omf, fid: opt.FID, lock: &sync.RWMutex{}}
}

// 初始化
func (ss *SSTable) Init() error {
	var ko *pb.BlockOffset
	var err error
	if ko, err = ss.initTable(); err != nil {
		return err
	}
	// init min key
	keyBytes := ko.GetKey()
	minKey := make([]byte, len(keyBytes))
	copy(minKey, keyBytes)
	ss.minKey = minKey

	blockLen := len(ss.idxTables.Offsets)
	ko = ss.idxTables.Offsets[blockLen-1]
	keyBytes = ko.GetKey()
	maxKey := make([]byte, 0)
	copy(maxKey, keyBytes)
	ss.maxKey = maxKey
	return nil
}

func (ss *SSTable) initTable() (bo *pb.BlockOffset, err error) {
	readPos := len(ss.f.Data)

	readPos -= 4
	buf := ss.readCheckError(readPos, 4)
	checksumLen := int(codec.BytesToU32(buf)) //得到 SSTable 文件中存储的校验和的长度 checksumLen
	if checksumLen < 0 {
		return nil, errors.New("checksum length less than zero. Data corrupted")
	}

	//
	readPos -= checksumLen
	expectedChk := ss.readCheckError(readPos, checksumLen) //然后再向前读取 checksumLen 个字节，读取期望的校验和 expectedChk。

	readPos -= 4
	buf = ss.readCheckError(readPos, 4)
	ss.idxLen = int(codec.BytesToU32(buf)) //获取索引表长度

	readPos -= ss.idxLen
	ss.idxStart = readPos
	data := ss.readCheckError(readPos, ss.idxLen) //SSTable 文件的 idxStart 位置向后读取 idxLen 个字节，这一段二进制数据存储的就是 SSTable 的索引表信息。
	if err := utils.VerifyChecksum(data, expectedChk); err != nil {
		return nil, errors.Wrapf(err, "failed to verify checksum for table: %s", ss.f.Fd.Name())
	}
	indexTable := &pb.TableIndex{}
	if err := proto.Unmarshal(data, indexTable); err != nil {
		return nil, err
		//数据反序列化为 TableIndex 类型
	}
	ss.idxTables = indexTable

	ss.hasBloomFilter = len(indexTable.BloomFilter) > 0
	//如果 indexTable 对象中存在偏移量，那么将返回第一个偏移量对象，否则返回一个错误。
	if len(indexTable.GetOffsets()) > 0 {
		return indexTable.GetOffsets()[0], nil
	}
	return nil, errors.New("read index fail, offset is nil")
}

func (ss *SSTable) Indexs() *pb.TableIndex {
	return ss.idxTables
}

func (ss *SSTable) MaxKey() []byte {
	return ss.maxKey
}

func (ss *SSTable) MinKey() []byte {
	return ss.minKey
}

func (ss *SSTable) FID() uint64 {
	return ss.fid
}

func (ss *SSTable) HasBloomFilter() bool {
	return ss.hasBloomFilter
}

func (ss *SSTable) read(off, sz int) ([]byte, error) {
	if len(ss.f.Data) > 0 {
		if len(ss.f.Data[off:]) < sz {
			return nil, io.EOF
		}
		return ss.f.Data[off : off+sz], nil
	}
	res := make([]byte, sz)
	_, err := ss.f.Fd.ReadAt(res, int64(off))
	return res, err
}

func (ss *SSTable) readCheckError(off, sz int) []byte {
	buf, err := ss.read(off, sz)
	utils.Panic(err)
	return buf
}

func (ss *SSTable) Bytes(off, sz int) ([]byte, error) {
	return ss.f.Bytes(off, sz)
}
