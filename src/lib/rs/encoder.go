package rs

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type encoder struct {
	writers []io.Writer
	enc     reedsolomon.Encoder
	cache   []byte
}

func NewEncoder(writers []io.Writer) *encoder {
	// reedsolomon.New生成了一个具有4个数据片+2个校验片的RS编码enc
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	//将输入参数 writers enc 作为生成的 encoder 结构体的成员返回
	return &encoder{writers, enc, nil}
}

func (e *encoder) Write(p []byte) (n int, err error) {
	length := len(p)
	current := 0
	// 将p待写入的数据以块的形式放入缓存
	for length != 0 {
		next := BLOCK_SIZE - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, p[current:current+next]...)

		// 如果缓存已满即调用Flush方法将缓存实际写入writers
		if len(e.cache) == BLOCK_SIZE {
			e.Flush()
		}
		current += next
		length -= next
	}
	return len(p), nil
}

func (e *encoder) Flush() {
	if len(e.cache) == 0 {
		return
	}
	// 将缓存的数据切成4个数据片
	shards, _ := e.enc.Split(e.cache)
	// 生成两个校验片
	e.enc.Encode(shards)

	// 将6个片数据依次写入writers并清空缓存
	for i := range shards {
		e.writers[i].Write(shards[i])
	}
	e.cache = []byte{}
}
