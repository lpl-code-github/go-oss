package temp

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"oss/src/lib/myLog"
	"strings"
)

func patch(w http.ResponseWriter, r *http.Request) {
	// 从请求url中获取获取uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]

	// 读取临时对象 信息文件
	tempinfo, e := readFromFile(uuid)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 临时文件存在

	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	// 临时对象数据文件
	datFile := infoFile + ".dat"

	// 打开临时对象数据文件
	f, e := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// 将请求正文写入临时对象数据文件
	_, e = io.Copy(f, r.Body)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 获取临时对象数据文件的信息 info
	info, e := f.Stat()
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 获取info的大小
	actual := info.Size()

	// 如果实际长度超过tempinfo记录的size
	if actual > tempinfo.Size {
		// 删除信息文件和临时数据文件
		os.Remove(datFile)
		os.Remove(infoFile)
		// 报错并返回500
		myLog.Error.Println("actual size", actual, "exceeds", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 读取相关信息文件
func readFromFile(uuid string) (*tempInfo, error) {
	// 打开临时对象 信息文件
	f, e := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	// 读取文件内容
	b, _ := ioutil.ReadAll(f)
	var info tempInfo

	// 反序列化为tempInfo结构体
	json.Unmarshal(b, &info)
	return &info, nil
}
