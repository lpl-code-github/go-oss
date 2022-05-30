package objects

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"oss/src/lib/utils"
	"strconv"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	// 获得对象名
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//unescape, _ := url.QueryUnescape(name)
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 获得版本号
	versionId := r.URL.Query()["version"]
	version := 0
	var e error
	if len(versionId) != 0 {
		version, e = strconv.Atoi(versionId[0])
		if e != nil {
			myLog.Error.Println(e)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	// 获得元数据
	meta, e := es.GetMetadata(bucket, name, version)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 得到stream流
	hash := url.PathEscape(meta.Hash)
	stream, e := GetStream(hash, meta.Size)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	offset := utils.GetOffsetFromHeader(r.Header)
	if offset != 0 {
		stream.Seek(offset, io.SeekCurrent)
		w.Header().Set("content-range", fmt.Sprintf("bytes %d-%d/%d", offset, meta.Size-1, meta.Size))
		w.WriteHeader(http.StatusPartialContent)
	}

	acceptGzip := false
	// 从header中检查Accept-Encoding，如果含有gzip，说明客户端可以接受gzip压缩数据
	encoding := r.Header["Accept-Encoding"]
	for i := range encoding {
		if encoding[i] == "gzip" {
			acceptGzip = true
			break
		}
	}

	// 设置响应头content-encoding为gzip
	if acceptGzip {
		w.Header().Set("content-encoding", "gzip")
		// 创建一个指向gzip.Writer结构体的指针w2
		w2 := gzip.NewWriter(w)
		// 将对象数据流stream写入w2
		io.Copy(w2, stream)
		w2.Close()
	} else {
		// 如果客户端不接受gzip，直接返回
		io.Copy(w, stream)
	}
	stream.Close()
	myLog.Info.Println(fmt.Sprintf("下载对象 桶：%s，对象名：%s，版本号：%d", bucket, name, version))
}
