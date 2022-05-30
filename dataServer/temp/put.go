package temp

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func put(w http.ResponseWriter, r *http.Request) {
	// 获取请求中的uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]

	// 读取相关信息文件
	tempinfo, e := readFromFile(uuid)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 打开本地temp目录下的信息文件
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	f, e := os.Open(datFile)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// 获取临时对象数据文件的信息 info
	info, e := f.Stat()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// info长度
	actual := info.Size()

	// 删除信息文件
	os.Remove(infoFile)

	// 如果实际长度 不等于 tempinfo记录的长度
	if actual != tempinfo.Size {
		// 删除临时文件
		os.Remove(datFile)
		log.Println("actual size mismatch, expect", tempinfo.Size, "actual", actual)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 临时数据文件转正
	commitTempObject(datFile, tempinfo)
}
