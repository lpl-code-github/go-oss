package objects

import (
	"compress/gzip"
	"io"
	"os"
	"oss/src/lib/myLog"
)

func sendFile(w io.Writer, file string) {
	f, e := os.Open(file)
	if e != nil {
		myLog.Error.Println(e)
		return
	}
	defer f.Close()

	// 在对象f上创建一个指向gzip.Reader的结构体指针gzipStream
	gzipStream, e := gzip.NewReader(f)
	if e != nil {
		myLog.Error.Println(e)
		return
	}

	// 读取gzipStream中的数据给w
	io.Copy(w, gzipStream)
	gzipStream.Close()
}
