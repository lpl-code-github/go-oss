package rs

import (
	"fmt"
	"io"
	"oss/src/lib/objectstream"
)

// RSPutStream 内嵌一个encoder结构体
type RSPutStream struct {
	*encoder
}

// NewRSPutStream 返回一个RSPutStream结构体
func NewRSPutStream(dataServers []string, hash string, size int64) (*RSPutStream, error) {
	// 检查dataServers数组长度是否等于6
	if len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("dataServers number mismatch")
	}

	// 计算每个分片大小
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS

	// 长度为6的io.Writer的数组
	writers := make([]io.Writer, ALL_SHARDS)
	var e error

	for i := range writers {
		writers[i], e = objectstream.NewTempPutStream(dataServers[i],
			fmt.Sprintf("%s.%d", hash, i), perShard)
		if e != nil {
			return nil, e
		}
	}
	// NewEncoder创建一个encoder指针enc
	enc := NewEncoder(writers)

	// 将enc作为RSPutStream的内嵌结构体返回
	return &RSPutStream{enc}, nil
}

func (s *RSPutStream) Commit(success bool) {
	// 调用encoder的Flush方法，将缓存中最多的数据写入
	s.Flush()

	// 对encoder的成员数组writers元素调用Commit方法，将六个临时对象依次转正或者删除
	for i := range s.writers {
		s.writers[i].(*objectstream.TempPutStream).Commit(success)
	}
}
