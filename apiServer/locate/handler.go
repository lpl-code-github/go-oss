package locate

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler 参数w：写入HTTP的响应，参数r：http请求对象
func Handler(w http.ResponseWriter, r *http.Request) {
	// r的成员Method变量记录了HTTP请求的方法
	m := r.Method

	// 检查请求方法
	if m != http.MethodGet { // 如果不是get请求，返回405
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 调用Locate 参数为文件名 进行定位
	info := Locate(strings.Split(r.URL.EscapedPath(), "/")[2])

	// 如果长度为0
	if len(info) == 0 {
		// 返回404
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 如果长度不为0 将info转为jsong格式
	b, _ := json.Marshal(info)

	// 响应体返回数据
	w.Write(b)
}
