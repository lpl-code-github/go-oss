package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"oss/src/lib/myLog"
	"strings"
)

func NodeSystemInfo(w http.ResponseWriter, r *http.Request) {
	// r的成员Method变量记录了HTTP请求的方法
	m := r.Method

	// 检查请求方法
	if m != http.MethodGet { // 如果不是get请求，返回405
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 从路径中获得节点ip
	nodeIp := strings.Split(r.URL.EscapedPath(), "/")[2]
	url := fmt.Sprintf("http://%s/systemInfo", nodeIp)
	if nodeIp == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	get, err := http.Get(url)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if get.StatusCode != http.StatusOK {
		log.Printf("从 %s 获取系统信息失败", nodeIp)
		w.WriteHeader(get.StatusCode)
		return
	}

	myLog.Info.Println(fmt.Sprintf("获取 %s 节点硬件信息", nodeIp))
	result, _ := ioutil.ReadAll(get.Body)
	w.Write(result)
}
