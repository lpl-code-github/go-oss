package objects

import (
	"net/http"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	// 获取url中对象hash值，作为参数调用getFile获得对象文件名file
	file := getFile(strings.Split(r.URL.EscapedPath(), "/")[2])

	if file == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 将对象内容输出到http响应
	sendFile(w, file)
}
