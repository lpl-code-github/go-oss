package temp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// 结构体tempInfo用于记录临时对象的uuid,名字、大小
type tempInfo struct {
	Uuid string
	Name string
	Size int64
}

func post(w http.ResponseWriter, r *http.Request) {
	// 随机生成uuid
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(output), "\n")

	// 获取对象名=散列值
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	// 获取头部的size=文件的大小
	size, e := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 拼接成tempInfo结构体
	t := tempInfo{uuid, name, size}
	// 创建临时文件对象信息的文件
	e = t.writeToFile()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 创建临时文件，这里是临时对象内容文件 <uuid>.dat
	os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid + ".dat")
	// 响应体返回uuid
	w.Write([]byte(uuid))
}

func (t *tempInfo) writeToFile() error {
	// 在temp目录下创建一个名为<uuid>的信息文件
	f, e := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid)
	if e != nil {
		return e
	}
	defer f.Close()

	// 将t转为json
	b, _ := json.Marshal(t)

	// 向<uuid>的临时文件写入b，这个文件是保存临时对象信息的
	f.Write(b)

	return nil
}
