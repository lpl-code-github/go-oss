package temp

import (
	"compress/gzip"
	"io"
	"net/url"
	"os"
	"oss/dataServer/locate"
	"oss/src/lib/utils"
	"strconv"
	"strings"
)

func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}

func commitTempObject(datFile string, tempinfo *tempInfo) {
	f, _ := os.Open(datFile)
	defer f.Close()
	d := url.PathEscape(utils.CalculateHash(f))
	f.Seek(0, io.SeekStart)
	// 创建正式对象文件w
	w, _ := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + tempinfo.Name + "." + d)
	// gzip.NewWriter创建压缩过后的w2
	w2 := gzip.NewWriter(w)
	// 将临时对象f中的数据复制进w2
	io.Copy(w2, f)
	w2.Close()
	// 删除临时对象文件
	os.Remove(datFile)
	// 添加对象定位缓存
	locate.Add(tempinfo.hash(), tempinfo.id())
}
