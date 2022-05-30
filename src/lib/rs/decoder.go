package rs

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type decoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64
	cache     []byte
	cacheSize int
	total     int64
}

func NewDecoder(readers []io.Reader, writers []io.Writer, size int64) *decoder {
	// 创建4+2RS码的解码器
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	// 设置decoder结构体中相应属性 并返回
	return &decoder{readers, writers, enc, size, nil, 0, 0}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	// 当cache中没有更多数据
	if d.cacheSize == 0 {
		// 调用getData方法
		e := d.getData()
		if e != nil {
			return 0, e
		}
	}

	length := len(p)
	// p的长度超出当前缓存的数据大小
	if d.cacheSize < length {
		// p的长度设置为缓存数据的大小
		length = d.cacheSize
	}

	d.cacheSize -= length
	// 将缓存中length长度的数据赋值给参数p
	copy(p, d.cache[:length])

	// 调整缓存只留下剩下的不慎
	d.cache = d.cache[length:]

	// 返回length，通知调用方本次读取一共有多少数据被复制到p中
	return length, nil
}

func (d *decoder) getData() error {
	// 判断当前已经解码的数组大小是否等于对象原始大小
	if d.total == d.size {
		// 相等 说明所有数据都已经被读取
		return io.EOF
	}

	// 长度为6的数组
	shards := make([][]byte, ALL_SHARDS)

	// 长度为0的整形数组
	repairIds := make([]int, 0)

	// 遍历6个shards
	for i := range shards {
		// 如果某个分片对应的reader为nil 说明分片丢失
		if d.readers[i] == nil {
			// repairIds中添加该分片的id
			repairIds = append(repairIds, i)
		} else {
			// shards[i]被初始化为一个8000的字节数组
			shards[i] = make([]byte, BLOCK_PER_SHARD)
			// 从reader中完整读取8000字节并保存在shards[i]中
			n, e := io.ReadFull(d.readers[i], shards[i])
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				// 如果发生了非EOF失败，shards[i]被设置为nil
				shards[i] = nil
			} else if n != BLOCK_PER_SHARD { // 如果读取的数组长度n 不到8000字节，将该shards的实际长度缩减为n
				shards[i] = shards[i][:n]
			}
		}
	}

	// 尝试将被置为nil的shards恢复出来
	e := d.enc.Reconstruct(shards)
	if e != nil {
		// 不可修复
		return e
	}
	// 如果修复成功
	for i := range repairIds {
		// 将需要恢复的分片数据写入相应的writer
		id := repairIds[i]
		d.writers[id].Write(shards[id])
	}

	// 遍历四个数据分片
	for i := 0; i < DATA_SHARDS; i++ {
		shardSize := int64(len(shards[i]))
		if d.total+shardSize > d.size {
			shardSize -= d.total + shardSize - d.size
		}
		// 将每个分片中数据添加到缓存cache中
		d.cache = append(d.cache, shards[i][:shardSize]...)
		// 修改缓存当前的大小
		d.cacheSize += int(shardSize)
		// 修改已经读取的全部数据的大小
		d.total += shardSize
	}
	return nil
}
