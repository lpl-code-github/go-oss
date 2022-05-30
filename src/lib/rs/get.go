package rs

import (
	"fmt"
	"io"
	"oss/src/lib/objectstream"
)

// RSGetStream 内嵌decoder结构体
type RSGetStream struct {
	*decoder
}

func NewRSGetStream(locateInfo map[int]string, dataServers []string, hash string, size int64) (*RSGetStream, error) {
	//	locateInfo+dataServers的总数是否 = 6，满足4+2RS码的需求
	if len(locateInfo)+len(dataServers) != ALL_SHARDS {
		// 不满足 返回错误
		return nil, fmt.Errorf("dataServers number mismatch")
	}

	// 创建一个长度为6的io.Reader数组 用于读取6个分片的数据
	readers := make([]io.Reader, ALL_SHARDS)

	// 遍历6个分片的id
	for i := 0; i < ALL_SHARDS; i++ {
		// 在locateInfo查找该分片所在的数据服务节点地址
		server := locateInfo[i]
		// 如果某个分片 id 相对的数据服务节点地址为空，分片丢失
		if server == "" {
			// 取一个随机数据服务节点补上
			locateInfo[i] = dataServers[0]
			dataServers = dataServers[1:]
			continue
		}

		// 如果服务数据节点存在 打开一个对象读取流用于读取该分片数据，打开的流被保存在reader数组中响应的元素中
		reader, e := objectstream.NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
		if e == nil {
			readers[i] = reader
		}
	}

	writers := make([]io.Writer, ALL_SHARDS)
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS
	var e error

	// 遍历readers
	for i := range readers {
		// 如果有个元素为nil
		if readers[i] == nil {
			// 创建相应的临时对象写入流用于恢复分片，打开的流被保存在writers数组
			writers[i], e = objectstream.NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), perShard)
			if e != nil {
				return nil, e
			}
		}
	}

	// 生成decoder结构体指针
	dec := NewDecoder(readers, writers, size)

	// 指针作为RSGetStream的内嵌结构体返回
	return &RSGetStream{dec}, nil
}

func (s *RSGetStream) Close() {
	// 遍历writers郑源，
	for i := range s.writers {
		// 如果某个分片的writers不为nil
		if s.writers[i] != nil {
			// 调用Commit将临时对象转正
			s.writers[i].(*objectstream.TempPutStream).Commit(true)
		}
	}
}

func (s *RSGetStream) Seek(offset int64, whence int) (int64, error) {
	// 只支持从当前位置起跳
	if whence != io.SeekCurrent {
		panic("only support SeekCurrent")
	}
	// 跳过的字节数不能<0
	if offset < 0 {
		panic("only support forward seek")
	}
	// 每次读取32000字节并丢弃，直到读到offset为止
	for offset != 0 {
		length := int64(BLOCK_SIZE)
		if offset < length {
			length = offset
		}
		buf := make([]byte, length)
		io.ReadFull(s, buf)
		offset -= length
	}
	return offset, nil
}
