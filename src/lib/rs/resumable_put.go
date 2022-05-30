package rs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"oss/src/lib/objectstream"
	"oss/src/lib/utils"

	"log"
	"net/http"
)

type resumableToken struct {
	Name    string
	Size    int64
	Hash    string
	Servers []string
	Uuids   []string
}

type RSResumablePutStream struct {
	*RSPutStream
	*resumableToken
}

// NewRSResumablePutStream 创建的stream是一个指向RSResumablePutStream的结构体指针
func NewRSResumablePutStream(dataServers []string, name, hash string, size int64) (*RSResumablePutStream, error) {
	// 创建一个RSPutStream类型的变量
	putStream, e := NewRSPutStream(dataServers, hash, size)
	if e != nil {
		return nil, e
	}

	// 从putStream的成员writers数组中获取6个分片的uuid
	uuids := make([]string, ALL_SHARDS)
	for i := range uuids {
		uuids[i] = putStream.writers[i].(*objectstream.TempPutStream).Uuid
	}

	// 创建resumableToken结构体token
	token := &resumableToken{name, size, hash, dataServers, uuids}

	// 将putStream和token作为RSResumablePutStream的成员返回
	return &RSResumablePutStream{putStream, token}, nil
}

func NewRSResumablePutStreamFromToken(token string) (*RSResumablePutStream, error) {
	// 对token做bash64解码
	b, e := base64.StdEncoding.DecodeString(token)
	if e != nil {
		return nil, e
	}

	var t resumableToken
	// 将json数据b 反序列化为resumableToken结构体t
	e = json.Unmarshal(b, &t)
	if e != nil {
		return nil, e
	}

	// 创建6个TempPutStream保存在数组中
	writers := make([]io.Writer, ALL_SHARDS)
	for i := range writers {
		writers[i] = &objectstream.TempPutStream{t.Servers[i], t.Uuids[i]}
	}

	// 以writers数组创建encoder结构体
	enc := NewEncoder(writers)

	// 以enc内嵌结构体创建RSPutStream，再以RSPutStream、t为内嵌结构体创建RSResumablePutStream结构体并返回
	return &RSResumablePutStream{&RSPutStream{enc}, &t}, nil
}

func (s *RSResumablePutStream) ToToken() string {
	// 将RSResumablePutStream转换为json
	b, _ := json.Marshal(s)
	// 返回bash64编码后的字符串
	return base64.StdEncoding.EncodeToString(b)
}

func (s *RSResumablePutStream) CurrentSize() int64 {
	// 发送head请求 获取第一个分片临时对象的大小
	r, e := http.Head(fmt.Sprintf("http://%s/temp/%s", s.Servers[0], s.Uuids[0]))
	if e != nil {
		log.Println(e)
		return -1
	}
	if r.StatusCode != http.StatusOK {
		log.Println(r.StatusCode)
		return -1
	}

	// 临时对象大小*4
	size := utils.GetSizeFromHeader(r.Header) * DATA_SHARDS
	if size > s.Size { // 超出对象大小
		// 返回对象大小
		size = s.Size
	}
	return size
}
